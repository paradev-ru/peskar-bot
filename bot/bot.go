package bot

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/leominov/peskar-bot/telegram"
)

const (
	EventsChannel = "jobs"
)

type Bot struct {
	telegram *telegram.Client
	redis    *RedisStore
	config   Config
}

func New(telegram *telegram.Client, config Config) *Bot {
	redis := NewRedis(config.RedisMaxIdle, config.RedisIdleTimeout, config.RedisAddr)
	return &Bot{
		telegram: telegram,
		config:   config,
		redis:    redis,
	}
}

func (b *Bot) SuccessReceived(result []byte) error {
	var job Job
	var err error
	if err = json.Unmarshal(result, &job); err != nil {
		return fmt.Errorf("Unmarshal error: %v (%s)", err, string(result))
	}
	for _, action := range b.config.Actions {
		reg, err := regexp.Compile(action.JobState)
		if err != nil {
			continue
		}
		if reg.FindString(job.State) == "" {
			continue
		}
		message, err := action.Template(job)
		if message == "" || err != nil {
			continue
		}
		if action.ChatId != "" {
			logrus.Debugf("%s. Send '%s' to %s", job.ID, message, action.ChatId)
			err = b.telegram.ChatSend(action.ChatId, message)
		} else {
			logrus.Debugf("%s. Send '%s' to default chat", job.ID, message)
			err = b.telegram.Send(message)
		}
		if err != nil {
			logrus.Error(err)
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
	sub := b.redis.NewSubscribe(EventsChannel)
	sub.SuccessReceivedCallback = b.SuccessReceived
	sub.RetryingPolicyCallback = b.RetryingPolicy
	return sub.Run()
}
