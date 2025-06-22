package configs

import (
	"github.com/omniful/go_commons/log"
)

var logger = log.DefaultLogger()

func GetLogger() *log.Logger {
	return logger
}