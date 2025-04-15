package sender

type EmailSender interface {
	Send(email, msg string) error
}
