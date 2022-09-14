package smtp

import (
	"gopkg.in/gomail.v2"
	"newTradingBot/models/users"
)

const mail = "support@infinance.app"
const password = "08091908Fekl@"
const host = "smtp.hostinger.com"
const port = 465

func SendPlainEmail(user users.User, subject, message string) error {
	sender := getDefaultSender()
	msg := gomail.NewMessage()

	msg.SetHeader("From", mail)
	msg.SetHeader("To", user.Email)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", message)

	return sender.DialAndSend(msg)
}

func getDefaultSender() *gomail.Dialer  {
	return gomail.NewDialer(host, port, mail, password)
}