package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/ak1m1tsu/barman/internal/app"
	"github.com/ak1m1tsu/barman/internal/infrastructure/config"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	configPath := flag.String("config", "configs/config.yaml", "Path to YAML config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		logrus.WithError(err).Fatal("config: failed to load")
	}

	a, err := app.New(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("app: failed to initialize")
	}
	defer func() {
		if err := a.Close(); err != nil {
			logrus.WithError(err).Error("app: failed to close")
		}
	}()

	if err := a.Run(); err != nil {
		logrus.WithError(err).Fatal("app: failed to run")
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logrus.Info("shutting down")
}
