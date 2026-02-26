package main

import (
	"flag"
	"fmt"
	"log"
	"net/smtp"
)

// go run . -u [gmail address] -P [gmail app password] -receiver [receiver's email address]
func main() {
	var senderAddress string
	var receiverAddress string
	var senderPassword string
	var smtpHost string
	var smtpPort int

	flag.StringVar(&senderAddress, "u", "", "Email address")
	flag.StringVar(&senderPassword, "P", "", "Password")
	flag.StringVar(&smtpHost, "h", "smtp.gmail.com", "SMTP host")
	flag.IntVar(&smtpPort, "p", 587, "SMTP port")
	flag.StringVar(&receiverAddress, "receiver", "", "Receiver's email address")
	flag.Parse()

	if receiverAddress == "" {
		log.Fatal("receiver email address is required")
	}

	mailSubject := "Hello, World!"
	mailBody := "This is a test email."

	client := NewEmailClient(senderAddress, senderPassword, smtpHost, smtpPort)
	err := client.Send(receiverAddress, mailSubject, mailBody)
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

func (c *EmailClient) Send(toAddress, subject, body string) error {
	message := c.composeMessage(toAddress, subject, body)
	return smtp.SendMail(c.smtpURL(), c.smtpAuth, c.address, []string{toAddress}, message)
}

func (c *EmailClient) smtpURL() string {
	return fmt.Sprintf("%s:%d", c.smtpHost, c.smtpPort)
}

func (c *EmailClient) composeMessage(toAddress, subject, body string) []byte {
	return fmt.Appendf(nil, "From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s\r\n", c.address, toAddress, subject, body)
}
