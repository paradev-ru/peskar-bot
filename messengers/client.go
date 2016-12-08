package messengers

type MessengerClient interface {
	Send(message string) error
	SendTo(to, message string) error
}
