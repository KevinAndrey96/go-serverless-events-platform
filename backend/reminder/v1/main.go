// Package main implements the Scheduled Event Reminder microservice handler
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/config"
	"github.com/KevinAndrey96/serverless-events-platform/backend/pkg/models"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// SMTPConfig holds credentials and connection parameters for the outbound mail server
type SMTPConfig struct {
	Host               string
	Port               string
	User               string
	Pass               string
	FromEmail          string
	InsecureSkipVerify bool
}

// loadSMTPConfig loads SMTP parameters from environment variables
func loadSMTPConfig() SMTPConfig {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "mail.events.instanceshape.com"
	}

	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "465"
	}

	user := os.Getenv("SMTP_USER")
	if user == "" {
		user = "no-responder@events.instanceshape.com"
	}

	pass := os.Getenv("SMTP_PASS")

	fromEmail := os.Getenv("FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = user
	}

	insecure := os.Getenv("SMTP_INSECURE_SKIP_VERIFY") == "true"

	return SMTPConfig{
		Host:               host,
		Port:               port,
		User:               user,
		Pass:               pass,
		FromEmail:          fromEmail,
		InsecureSkipVerify: insecure,
	}
}

// Handler processes the scheduled reminder invocation
func Handler(ctx context.Context) error {
	log.Printf("[INFO] Starting event reminder worker execution")
	smtpCfg := loadSMTPConfig()

	db := config.GetDynamoDBClient()
	tableName := config.GetTableName()

	queryInput := &dynamodb.QueryInput{
		TableName: aws.String(tableName),
		IndexName: aws.String("sk-index"),
		KeyConditions: map[string]*dynamodb.Condition{
			"sk": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String("METADATA#USER"),
					},
				},
			},
		},
	}

	result, err := db.QueryWithContext(ctx, queryInput)
	if err != nil {
		log.Printf("[ERROR] Failed to query user attendees: %v", err)
		return err
	}

	var attendees []models.UserAttendee
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &attendees)
	if err != nil {
		log.Printf("[ERROR] Failed to unmarshal attendees list: %v", err)
		return err
	}

	log.Printf("[INFO] Retrieved %d attendees to evaluate for upcoming reminders", len(attendees))

	loc := time.FixedZone("COT", -5*60*60)
	now := time.Now().In(loc)
	layout := "2006-01-02 15:04:05"

	sentCount := 0
	for _, attendee := range attendees {
		if attendee.Email == "" || attendee.EventID == "" {
			continue
		}

		eventItem, err := fetchEventDetails(ctx, db, tableName, attendee.EventID)
		if err != nil {
			log.Printf("[WARN] Unable to fetch event %s for attendee %s: %v", attendee.EventID, attendee.Email, err)
			continue
		}

		eventTimeString := fmt.Sprintf("%s %s:00", eventItem.Date, eventItem.Hour)
		eventTime, err := time.ParseInLocation(layout, eventTimeString, loc)
		if err != nil {
			log.Printf("[WARN] Unable to parse event date/time '%s': %v", eventTimeString, err)
			continue
		}

		diff := eventTime.Sub(now)
		hoursRemaining := diff.Hours()

		// Send reminder if event starts between 0 and 24 hours from now
		if hoursRemaining > 0 && hoursRemaining <= 24 {
			log.Printf("[INFO] Event '%s' starts in %.1f hours. Sending reminder to %s", eventItem.Name, hoursRemaining, attendee.Email)
			err = sendReminderEmail(smtpCfg, attendee.Email, eventItem.Name, eventItem.Date, eventItem.Hour)
			if err != nil {
				log.Printf("[ERROR] Failed sending email to %s: %v", attendee.Email, err)
			} else {
				sentCount++
			}
		}
	}

	log.Printf("[INFO] Reminder cycle finished. Sent %d notification emails", sentCount)
	return nil
}

// fetchEventDetails loads the parent event metadata from DynamoDB
func fetchEventDetails(ctx context.Context, db *dynamodb.DynamoDB, tableName, eventID string) (*models.Event, error) {
	pk := eventID
	if !strings.HasPrefix(pk, "Event#") {
		pk = "Event#" + pk
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {
				S: aws.String(pk),
			},
			"sk": {
				S: aws.String("METADATA#EVENT"),
			},
		},
	}

	result, err := db.GetItemWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, fmt.Errorf("event item with pk=%s not found", pk)
	}

	var eventItem models.Event
	err = dynamodbattribute.UnmarshalMap(result.Item, &eventItem)
	if err != nil {
		return nil, err
	}

	return &eventItem, nil
}

// sendReminderEmail handles authenticated TLS SMTP transmission
func sendReminderEmail(cfg SMTPConfig, recipient, eventName, eventDate, eventHour string) error {
	if cfg.Pass == "" {
		log.Printf("[WARN] SMTP_PASS not configured in environment. Skipping email dispatch to %s (dry-run)", recipient)
		return nil
	}

	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		ServerName:         cfg.Host,
	}

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls connection failed to %s: %w", addr, err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client creation failed: %w", err)
	}
	defer client.Quit()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth failed: %w", err)
	}

	if err := client.Mail(cfg.FromEmail); err != nil {
		return fmt.Errorf("mail from failed: %w", err)
	}

	if err := client.Rcpt(recipient); err != nil {
		return fmt.Errorf("rcpt to failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data writer failed: %w", err)
	}

	subject := fmt.Sprintf("Recordatorio: Tu evento '%s' es mañana", eventName)
	body := fmt.Sprintf("Hola!\n\nTe recordamos que tu evento '%s' está programado para el día %s a las %s.\n\nPor favor ten a mano tu código QR para el ingreso.\n\nSaludos,\nEquipo de Eventos", eventName, eventDate, eventHour)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", cfg.FromEmail, recipient, subject, body)

	if _, err = w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("writing body failed: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("closing writer failed: %w", err)
	}

	log.Printf("[INFO] Reminder email successfully delivered to %s", recipient)
	return nil
}

func main() {
	lambda.Start(Handler)
}
