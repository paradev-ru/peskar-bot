package bot

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/leominov/peskar-bot/messengers"
)

const (
	JobEventsChannel = "peskar.job.events"
	JobLogChannel    = "peskar.job.logs"
)

type Bot struct {
	Name   string
	client messengers.MessengerClient
	redis  *RedisStore
	config Config
}

func New(name string, client messengers.MessengerClient, config Config) *Bot {
	redis := NewRedis(config.RedisMaxIdle, config.RedisIdleTimeout, config.RedisAddr)
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "na"
	}
	return &Bot{
		Name:   fmt.Sprintf("%s-%s", name, hostname),
		client: client,
		config: config,
		redis:  redis,
	}
}

func (b *Bot) Log(jobID, message string) error {
	l := JobLog{
		Initiator: b.Name,
		JobID:     jobID,
		Message:   message,
	}
	return b.redis.Send(JobLogChannel, l)
}

func (b *Bot) SuccessReceived(result []byte) error {
	var job JobEntry
	var err error
	if err = json.Unmarshal(result, &job); err != nil {
		return fmt.Errorf("Unmarshal error: %v (%s)", err, string(result))
	}
	for _, action := range b.config.Actions {
		if regexp.MustCompile(action.JobState).FindString(job.State) == "" {
			continue
		}
		message, err := action.Template(job)
		if message == "" || err != nil {
			continue
		}
		if action.ChatId != "" {
			logrus.Debugf("%s. Send '%s' to %s", job.ID, message, action.ChatId)
			err = b.client.SendTo(action.ChatId, message)
		} else {
			logrus.Debugf("%s. Send '%s' to default chat", job.ID, message)
			err = b.client.Send(message)
		}
		if err != nil {
			logrus.Error(err)
			b.Log(job.ID, err.Error())
		}
	}
	return nil
}

func (b *Bot) RetryingPolicy(attempts int, duration time.Duration) error {
	logrus.Infof("Wait Redis for a 10 seconds (#%d, %v)", attempts, duration)
	time.Sleep(10 * time.Second)
	return nil
}

func (b *Bot) Validate() error {
	if err := b.redis.Check(); err != nil {
		return fmt.Errorf("Error creating redis connection: %+v", err)
	}
	return nil
}

func (b *Bot) Process() error {
	logrus.Info("Waiting for incoming events...")
	sub := b.redis.NewSubscribe(JobEventsChannel)
	sub.SuccessReceivedCallback = b.SuccessReceived
	sub.RetryingPolicyCallback = b.RetryingPolicy
	return sub.Run()
}
