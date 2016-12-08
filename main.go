package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/leominov/peskar-bot/bot"
	"github.com/leominov/peskar-bot/messengers"
	"github.com/leominov/peskar-bot/messengers/empty"
	"github.com/leominov/peskar-bot/messengers/telegram"
)

const (
	BaseName = "peskar-bot"
)

var (
	b               *bot.Bot
	messengerClient messengers.MessengerClient
)

func main() {
	flag.Parse()
	if printVersion {
		fmt.Printf("%s %s\n", BaseName, Version)
		os.Exit(0)
	}

	if err := initConfig(); err != nil {
		logrus.Fatal(err)
	}

	doneChan := make(chan bool)

	if telegramConfig.Enabled {
		messengerClient = telegram.New(telegramConfig)
	} else {
		messengerClient = empty.New()
	}

	logrus.Infof("Starting %s", BaseName)
	b = bot.New(BaseName, messengerClient, botConfig)
	if err := b.Validate(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	go b.Process()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-signalChan:
			logrus.Infof("Captured %v. Exiting...", s)
			close(doneChan)
		case <-doneChan:
			os.Exit(0)
		}
	}
}
