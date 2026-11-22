package models

// Event represents an event entity in the DynamoDB single-table schema
type Event struct {
	PK            string  `json:"pk" dynamodbav:"pk"`
	SK            string  `json:"sk" dynamodbav:"sk"`
	Name          string  `json:"name" dynamodbav:"name"`
	Capacity      string  `json:"capacity" dynamodbav:"capacity"`
	Price         string  `json:"price" dynamodbav:"price"`
	Date          string  `json:"date" dynamodbav:"date"`
	Hour          string  `json:"hour" dynamodbav:"hour"`
	Status        string  `json:"status" dynamodbav:"status"`
	Bank          *string `json:"bank,omitempty" dynamodbav:"bank,omitempty"`
	Account       *string `json:"account,omitempty" dynamodbav:"account,omitempty"`
	Lastname      *string `json:"lastname,omitempty" dynamodbav:"lastname,omitempty"`
	Gender        *string `json:"gender,omitempty" dynamodbav:"gender,omitempty"`
	Email         *string `json:"email,omitempty" dynamodbav:"email,omitempty"`
	Phone         *string `json:"phone,omitempty" dynamodbav:"phone,omitempty"`
	EventID       *string `json:"eventid,omitempty" dynamodbav:"eventid,omitempty"`
	UserID        *string `json:"userid,omitempty" dynamodbav:"userid,omitempty"`
	Scholarship   *string `json:"scholarship,omitempty" dynamodbav:"scholarship,omitempty"`
	PaymentStatus *string `json:"paymentstatus,omitempty" dynamodbav:"paymentstatus,omitempty"`
}

// UserAttendee represents an attendee registered for an event
type UserAttendee struct {
	PK            string `json:"pk" dynamodbav:"pk"`
	SK            string `json:"sk" dynamodbav:"sk"`
	Name          string `json:"name" dynamodbav:"name"`
	Lastname      string `json:"lastname" dynamodbav:"lastname"`
	Phone         string `json:"phone" dynamodbav:"phone"`
	Date          string `json:"date" dynamodbav:"date"`
	Status        string `json:"status" dynamodbav:"status"`
	EventID       string `json:"eventid" dynamodbav:"eventid"`
	Scholarship   string `json:"scholarship" dynamodbav:"scholarship"`
	UserID        string `json:"userid" dynamodbav:"userid"`
	PaymentStatus string `json:"paymentstatus" dynamodbav:"paymentstatus"`
	Email         string `json:"email,omitempty" dynamodbav:"email,omitempty"`
}

// UpdateStatusRequest payload for updating event lifecycle status
type UpdateStatusRequest struct {
	PK     string `json:"pk"`
	SK     string `json:"sk"`
	Status string `json:"status"`
}

// UpdatePaymentStatusRequest payload for updating payment status
type UpdatePaymentStatusRequest struct {
	PK            string `json:"pk"`
	SK            string `json:"sk"`
	PaymentStatus string `json:"paymentstatus"`
	Status        string `json:"status,omitempty"`
}

// GetEffectiveStatus returns the payment status regardless of whether 'paymentstatus' or 'status' key was sent
func (r *UpdatePaymentStatusRequest) GetEffectiveStatus() string {
	if r.PaymentStatus != "" {
		return r.PaymentStatus
	}
	return r.Status
}

// ErrorResponse standard error structure for API responses
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
