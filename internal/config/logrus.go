package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewLogger(v *viper.Viper) *logrus.Logger {

	log := logrus.New()
	log.SetFormatter(
		&logrus.JSONFormatter{},
	)

	level := v.GetInt("log.Level")
	if level == 0 {
		level = int(logrus.InfoLevel)
	}

	log.SetLevel(logrus.InfoLevel)

	return log

}
