package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/leominov/peskar-bot/bot"
	"github.com/leominov/peskar-bot/telegram"
)

var (
	b              *bot.Bot
	telegramClient *telegram.Client
)

func main() {
	flag.Parse()
	if printVersion {
		fmt.Printf("peskar-bot %s\n", Version)
		os.Exit(0)
	}

	logrus.Info("Starting peskar-bot")

	if err := initConfig(); err != nil {
		logrus.Fatal(err)
	}

	doneChan := make(chan bool)

	if telegramConfig.Enabled {
		telegramClient = telegram.New(telegramConfig)
	}

	b = bot.New(telegramClient, botConfig)
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
