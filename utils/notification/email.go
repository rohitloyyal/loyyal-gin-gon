package notification

import (
	"os"

	"gopkg.in/gomail.v2"
)

func SendEmailNotification(subject string, to string) error {
	m := gomail.NewMessage()

	m.SetHeader("From", os.Getenv("SMTP_EMAIL_FROM"))
	m.SetHeader("To", to)
	// m.SetHeader("To", "rohit@loyyal.com", "rohit@loyyal.com")
	// m.SetAddressHeader("Cc", "rohitroyrr8@gmail.com", "Dan")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")
	// m.Attach("/home/Alex/lolcat.jpg")

	d := gomail.NewDialer(os.Getenv("SMTP_EMAIL_HOST"), 587, os.Getenv("SMTP_EMAIL_USERNAME"), os.Getenv("SMTP_EMAIL_PASSWORD"))

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}

	return nil
}
