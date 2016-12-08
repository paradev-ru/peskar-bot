package empty

import "github.com/Sirupsen/logrus"

type Client struct {
}

func New() *Client {
	return &Client{}
}

func (c *Client) Send(message string) error {
	logrus.Infof("Mute message '%s'", message)
	return nil
}

func (c *Client) SendTo(to, message string) error {
	logrus.Infof("Mute message '%s' to %s", message, to)
	return nil
}
