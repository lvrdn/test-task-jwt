package sender

type EmailSender interface {
	Send(guid, msg string) error
}
