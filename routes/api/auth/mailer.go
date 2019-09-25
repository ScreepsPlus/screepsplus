package auth

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/volatiletech/authboss"
)

// SESMailer implements authboass email sending via AWS SES
type SESMailer struct {
	sess *session.Session
	ses  *ses.SES
}

// NewSESMailer creates a new SESMailer
func NewSESMailer() *SESMailer {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		panic(fmt.Errorf("Error creating SES Mailer: %v", err))
	}
	ses := ses.New(sess)

	return &SESMailer{
		sess: sess,
		ses:  ses,
	}
}

// Send an e-mail
func (s *SESMailer) Send(ctx context.Context, mail authboss.Email) error {
	if len(mail.TextBody) == 0 && len(mail.HTMLBody) == 0 {
		return errors.New("refusing to send mail without text or html body")
	}
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses:  mapAddrs(mail.To, mail.ToNames),
			CcAddresses:  mapAddrs(mail.Cc, mail.CcNames),
			BccAddresses: mapAddrs(mail.Bcc, mail.BccNames),
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(mail.HTMLBody),
				},
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(mail.TextBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(mail.Subject),
			},
		},
		Source: mapAddr(mail.From, mail.FromName),
	}
	_, err := s.ses.SendEmail(input)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}
	return nil
}

func mapAddr(addr string, name string) *string {
	if len(name) == 0 {
		return aws.String(addr)
	}
	return aws.String(fmt.Sprintf("%s <%s>", name, addr))
}

func mapAddrs(addrs []string, names []string) []*string {
	cnt := len(addrs)
	ret := make([]*string, cnt)
	if len(names) == 0 {
		for i, a := range addrs {
			ret[i] = aws.String(a)
		}
	} else {
		for i, a := range addrs {
			ret[i] = mapAddr(a, names[i])
		}
	}
	return ret
}
