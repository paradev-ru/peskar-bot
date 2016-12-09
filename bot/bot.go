package bot

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/leominov/peskar-bot/messengers"
	"github.com/leominov/peskar-hub/peskar"
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
	l := peskar.LogItem{
		Initiator: b.Name,
		JobID:     jobID,
		Message:   message,
	}
	logrus.Infof("%s: %s", jobID, message)
	return b.redis.Send(peskar.JobLogChannel, l)
}

func (b *Bot) SuccessReceived(result []byte) error {
	var job peskar.Job
	var err error
	var counter int
	if err = json.Unmarshal(result, &job); err != nil {
		return fmt.Errorf("Unmarshal error: %v (%s)", err, string(result))
	}
	for _, notify := range b.config.NotifyList {
		if regexp.MustCompile(notify.JobState).FindString(job.State) == "" {
			continue
		}
		message, err := notify.Template(job)
		if message == "" || err != nil {
			continue
		}
		counter++
		if counter == 1 {
			b.Log(job.ID, "Got a job")
		}
		if notify.ChatId != "" {
			b.Log(job.ID, fmt.Sprintf("Sending message to %s (%s)...", b.client.GetName(), notify.ChatId))
			err = b.client.SendTo(notify.ChatId, message)
		} else {
			b.Log(job.ID, fmt.Sprintf("Sending message to %s...", b.client.GetName()))
			err = b.client.Send(message)
		}
		if err != nil {
			b.Log(job.ID, err.Error())
			continue
		}
	}
	if counter > 0 {
		b.Log(job.ID, "Done")
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
	sub := b.redis.NewSubscribe(peskar.JobEventsChannel)
	sub.SuccessReceivedCallback = b.SuccessReceived
	sub.RetryingPolicyCallback = b.RetryingPolicy
	return sub.Run()
}
