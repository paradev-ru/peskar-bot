package messengers

type MessengerClient interface {
	GetName() string
	Send(message string) error
	SendTo(to, message string) error
}
