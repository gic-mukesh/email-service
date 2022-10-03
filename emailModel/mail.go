package mail

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Mail struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	MailSendTo  []string           `bson:"to,omitempty" json:"to,omitempty"`
	MailSendCC  []string           `bson:"cc,omitempty" json:"cc,omitempty"`
	MailSendBCC []string           `bson:"bcc,omitempty" json:"bcc,omitempty"`
	MailSubject []string           `bson:"subject,omitempty" json:"subject,omitempty"`
	MailBody    *Body              `bson:"body,omitempty" json:"body,omitempty"`
	SenderName  string             `bson:"sender,omitempty" json:"sender,omitempty"`
	Attachments []string           `bson:"attachments,omitempty" json:"attachments,omitempty"`
	Time        time.Time
}

type Body struct {
	Salutation string `bson:"salutation,omitempty" json:"salutation,omitempty"`
	Message    string `bson:"message,omitempty" json:"message,omitempty"`
	Closing    string `bson:"closing,omitempty" json:"closing,omitempty"`
}

type Search struct {
	MailSendTo  string `bson:"to,omitempty" json:"to,omitempty"`
	MailSendCC  string `bson:"cc,omitempty" json:"cc,omitempty"`
	MailSendBCC string `bson:"bcc,omitempty" json:"bcc,omitempty"`
	MailSubject string `bson:"subject,omitempty" json:"subject,omitempty"`
	Date        string `bson:"date,omitempty" json:"date,omitempty"`
}
