package event

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"strings"
)

const (
	ticketSeparator = "."
	eventIDBytes    = 10
	attendeeIDBytes = 16
)

// NewEventID returns a short random identifier that is safe in URLs.
func NewEventID() string {
	return strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes(eventIDBytes)))
}

// NewAttendeeID returns 128 random bits; the ticket is only as hard to forge as this value.
func NewAttendeeID() string {
	return base64.RawURLEncoding.EncodeToString(randomBytes(attendeeIDBytes))
}

// ParseTicket splits a QR value into event and attendee identifiers.
func ParseTicket(ticket string) (string, string, error) {
	eventID, attendeeID, found := strings.Cut(strings.TrimSpace(ticket), ticketSeparator)
	if !found || eventID == "" || len(attendeeID) != base64.RawURLEncoding.EncodedLen(attendeeIDBytes) {
		return "", "", ErrInvalidTicket
	}
	if _, err := base64.RawURLEncoding.DecodeString(attendeeID); err != nil {
		return "", "", ErrInvalidTicket
	}

	return eventID, attendeeID, nil
}

func randomBytes(n int) []byte {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)

	return buf
}
