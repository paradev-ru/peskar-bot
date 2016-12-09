package empty

import "github.com/Sirupsen/logrus"

type Client struct {
	Name string
}

func New() *Client {
	return &Client{
		Name: "Empty",
	}
}

func (c *Client) GetName() string {
	return c.Name
}

func (c *Client) Send(message string) error {
	logrus.Infof("Mute message '%s'", message)
	return nil
}

func (c *Client) SendTo(to, message string) error {
	logrus.Infof("Mute message '%s' to %s", message, to)
	return nil
}
