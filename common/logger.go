package common


import (
	"github.com/auto-cdp/utils/log"
)

func InitLog() {
	config := Config().Log
	log.InitLog(config.Filename, config.MaxSize, config.MaxBackups, config.MaxAge, config.Compress)
	log.SetLevel(config.LogLevel)
}

