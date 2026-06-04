package utils

import "github.com/resend/resend-go/v3"

type EmailParams struct {
	To      []string
	From    string
	Subject string
	Content string
}

type EmailSender struct {
	ResendKey string
}

func NewEmailSender(resendKey string) *EmailSender {
	return &EmailSender{
		ResendKey: resendKey,
	}
}

func (s *EmailSender) SendEmail(params *EmailParams) error {
	client := resend.NewClient(s.ResendKey)

	request := &resend.SendEmailRequest{
		From:    params.From,
		To:      params.To,
		Subject: params.Subject,
		Html:    params.Content,
	}

	if _, err := client.Emails.Send(request); err != nil {
		return err
	}

	return nil
}
