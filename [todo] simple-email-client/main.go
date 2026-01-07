package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/smtp"
)

func main() {
	var senderAddress string
	var senderPassword string
	var smtpHost string
	var smtpPort int

	flag.StringVar(&senderAddress, "u", "", "Email address")
	flag.StringVar(&senderPassword, "P", "", "Password")
	flag.StringVar(&smtpHost, "h", "", "SMTP host")
	flag.IntVar(&smtpPort, "p", 587, "SMTP port")
	flag.Parse()

	receiverAddress := "receiver@example.com"
	mailSubject := "Hello, World!"
	mailBody := "This is a test email."

	globalCtx := context.Background()

	client := NewEmailClient(senderAddress, senderPassword, smtpHost, smtpPort)
	err := client.Send(globalCtx, receiverAddress, mailSubject, mailBody)
	if err != nil {
		log.Fatalf("error sending new email: %s", err)
	}

}

type EmailClient struct {
	address  string
	smtpHost string
	smtpPort int
	smtpAuth smtp.Auth
}

func NewEmailClient(address, password, smtpHost string, smtpPort int) *EmailClient {
	smtpAuth := smtp.PlainAuth("", address, password, smtpHost)

	return &EmailClient{
		address:  address,
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		smtpAuth: smtpAuth,
	}
}

func (c *EmailClient) Send(ctx context.Context, toAddress, subject, body string) error {
	message := c.composeMessage(toAddress, subject, body)

	return smtp.SendMail(c.smtpURL(), c.smtpAuth, c.address, []string{toAddress}, message)
}

func (c *EmailClient) smtpURL() string {
	return fmt.Sprintf("%s:%d", c.smtpHost, c.smtpPort)
}

func (c *EmailClient) composeMessage(toAddress, subject, body string) []byte {
	return fmt.Appendf(nil, "From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s\r\n", c.address, toAddress, subject, body)
}
