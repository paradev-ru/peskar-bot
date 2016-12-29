package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Sirupsen/logrus"
	"github.com/paradev-ru/peskar-bot/bot"
	"github.com/paradev-ru/peskar-bot/messengers/telegram"
)

const (
	DefaultConfigFile       = "/opt/peskar/peskar-bot.toml"
	DefaultRedisAddr        = "redis://localhost:6379/0"
	DefaultRedisIdleTimeout = 240 * time.Second
	DefaultRedisMaxIdle     = 3
)

var (
	config           Config
	logLevel         string
	telegramConfig   telegram.Config
	botConfig        bot.Config
	redisAddr        string
	redisIdleTimeout time.Duration
	redisMaxIdle     int
	configFile       string
	printVersion     bool
	telegramToken    string
)

type Config struct {
	Telegram         telegram.Config `toml:"telegram"`
	NotifyList       []*bot.Notify   `toml:"notify"`
	RedisAddr        string          `toml:"-"`
	RedisIdleTimeout time.Duration   `toml:"-"`
	RedisMaxIdle     int             `toml:"-"`
	LogLevel         string          `toml:"-"`
}

func init() {
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.StringVar(&configFile, "config-file", "", "the bot config file")
	flag.StringVar(&logLevel, "log-level", "", "level which bot should log messages")
	flag.StringVar(&redisAddr, "redis-addr", "", "Redis server URL")
	flag.DurationVar(&redisIdleTimeout, "redis-idle-timeout", 0*time.Second, "close Redis connections after remaining idle for this duration")
	flag.IntVar(&redisMaxIdle, "redis-max-idle", 0, "Maximum number of idle connections in the Redis pool")
}

func initConfig() error {
	if configFile == "" {
		if _, err := os.Stat(DefaultConfigFile); !os.IsNotExist(err) {
			configFile = DefaultConfigFile
		}
	}

	config = Config{
		Telegram:         telegram.NewConfig(),
		RedisAddr:        DefaultRedisAddr,
		RedisIdleTimeout: DefaultRedisIdleTimeout,
		RedisMaxIdle:     DefaultRedisMaxIdle,
	}

	if configFile == "" {
		logrus.Info("Skipping peskar-bot config file.")
	} else {
		logrus.Info("Loading " + configFile)
		configBytes, err := ioutil.ReadFile(configFile)
		if err != nil {
			return err
		}
		_, err = toml.Decode(string(configBytes), &config)
		if err != nil {
			return err
		}
	}

	processEnv()

	processFlags()

	if config.LogLevel != "" {
		level, err := logrus.ParseLevel(config.LogLevel)
		if err != nil {
			return err
		}
		logrus.SetLevel(level)
	}

	if config.RedisAddr == "" {
		return errors.New("Must specify Redis server URL using -redis-addr")
	}

	if config.RedisIdleTimeout == 0*time.Second {
		return errors.New("Must specify Redis idle timeout using -redis-idle-timeout")
	}

	if config.RedisMaxIdle == 0 {
		return errors.New("Must specify Redis max idle using -redis-max-idle")
	}

	if len(config.NotifyList) == 0 {
		return errors.New("Notify list cant be empty, check config file")
	}

	if config.Telegram.Enabled && config.Telegram.Token == "" {
		return errors.New("Telegram enabled. Must specify token")
	}

	for id, notify := range config.NotifyList {
		if notify.JobState == "" {
			return fmt.Errorf("Must specify notify[%d].job_state", id)
		}
		if _, err := regexp.Compile(notify.JobState); err != nil {
			return fmt.Errorf("%v in notify[%d].job_state", err, id)
		}
		if notify.Message == "" {
			return fmt.Errorf("Must specify notify[%d].message", id)
		}
		if config.Telegram.Enabled {
			if config.Telegram.ChatId == "" && notify.ChatId == "" {
				return fmt.Errorf("Must specify notify[%d].chat_id or telegram.chat_id", id)
			}
		}
	}

	telegramConfig = config.Telegram
	botConfig = bot.Config{
		NotifyList:       config.NotifyList,
		RedisAddr:        config.RedisAddr,
		RedisIdleTimeout: config.RedisIdleTimeout,
		RedisMaxIdle:     config.RedisMaxIdle,
	}

	return nil
}

func processEnv() {
	redisAddrEnv := os.Getenv("PESKAR_BOT_REDIS_ADDR")
	if len(redisAddrEnv) > 0 {
		config.RedisAddr = redisAddrEnv
	}
	tokenEnd := os.Getenv("PESKAR_BOT_TELEGRAM_TOKEN")
	if len(tokenEnd) > 0 {
		config.Telegram.Token = tokenEnd
	}
}

func processFlags() {
	flag.Visit(setConfigFromFlag)
}

func setConfigFromFlag(f *flag.Flag) {
	switch f.Name {
	case "telegram-token":
		config.Telegram.Token = telegramToken
	case "redis-addr":
		config.RedisAddr = redisAddr
	case "redis-idle-timeout":
		config.RedisIdleTimeout = redisIdleTimeout
	case "redis-max-idle":
		config.RedisMaxIdle = redisMaxIdle
	case "log-level":
		config.LogLevel = logLevel
	}
}
