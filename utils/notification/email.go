package notification

import (
	"gopkg.in/gomail.v2"
)

func SendEmailNotification() error {
	// auth := smtp.PlainAuth("", "apikey", "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", "smtp.sendgrid.net")

	// from := "no-reply@loyyal.net"
	// to := []string{"rohit@loyyal.com"}

	// msg := []byte("To: rohit@loyyal.com\r\n" +
	// 	"Subject: Why aren’t you using Mailtrap yet?\r\n" +
	// 	"\r\n" +
	// 	"Here’s the space for our great sales pitch\r\n")

	// err := smtp.SendMail("smtp.sendgrid.net:587", auth, from, to, msg)

	// if err != nil {
	// 	return err
	// }

	// return nil

	m := gomail.NewMessage()
	m.SetHeader("From", "no-reply@loyyal.net")
	m.SetHeader("To", "rohit@loyyal.com", "rohit@loyyal.com")
	m.SetAddressHeader("Cc", "rohitroyrr8@gmail.com", "Dan")
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")
	// m.Attach("/home/Alex/lolcat.jpg")

	d := gomail.NewDialer("smtp.sendgrid.net", 587, "apikey", "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")

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}

	return nil
}
