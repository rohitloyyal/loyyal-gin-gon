package notification

import (
	"gopkg.in/gomail.v2"
)

func SendEmailNotification() error {
	m := gomail.NewMessage()
	m.SetHeader("From", SMTP_FROM)
	m.SetHeader("To", "rohit@loyyal.com", "rohit@loyyal.com")
	m.SetAddressHeader("Cc", "rohitroyrr8@gmail.com", "Dan")
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")
	// m.Attach("/home/Alex/lolcat.jpg")

	d := gomail.NewDialer(SMTP_HOST, SMTP_PORT, SMTP_USERNAME , SMTP_PASSWORD)

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}

	return nil
}
