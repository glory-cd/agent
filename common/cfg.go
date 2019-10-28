package common

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

type RedisConfig struct {
	Host      string        `json:"host"`
	MaxIdle   int           `json:"maxidele"`
	MaxActive int           `json:"maxactive"`
	Timeout   time.Duration `json:"timeout"`
}

type LogConfig struct {
	LogLevel   string `json:"loglevel"`   // log level
	Filename   string `json:"filename"`   // log file path
	MaxSize    int    `json:"maxsize"`    // max size：M
	MaxBackups int    `json:"maxbackups"` // How many backups of log files can be saved at most
	MaxAge     int    `json:"maxage"`     // How many days should the file be kept
	Compress   bool   `json:"compress"`   // Whether to enable compression
}

type Rest struct {
	Addr string `json:"addr"`
}

type GlobalConfig struct {
	Debug      bool         `json:"debug"`
	Redis      *RedisConfig `json:"redis"`
	Rest       *Rest        `json:"rest"`
	FileServer *StoreServer `json:"storeserver"`
	Log        *LogConfig   `json:"log"`
}

type StoreServer struct {
	Addr     string `json:"addr"`
	Type     string `json:"type"`
	UserName string `json:"username"`
	PassWord string `json:"password"`
	S3Region string `json:"s3region"`
	S3Bucket string `json:"s3bucket"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	lock       = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	lock.RLock()
	defer lock.RUnlock()
	return config
}

// Get the service configuration from the etcd and initialize it
func InitConfig() {

	configContent, err := EtcdClient.Get(ConfigKey, false)

	log.Printf("ConfigKey:%+v", configContent)

	if err != nil {
		log.Fatal("get config faild:", "fail:", err)
	}

	if len(configContent) == 0 {
		NewConfigInEtcd(&configContent)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent[ConfigKey]), &c)
	if err != nil {
		log.Fatal("parse config :", "fail:", err)
	}

	lock.Lock()
	defer lock.Unlock()

	config = &c

}

// Create a new configuration in the etcd
func NewConfigInEtcd(content *map[string]string) {
	temp, err := EtcdClient.Get(ConfigTemplate, false)

	if err != nil {
		log.Fatal("get config template faild:", "fail:", err)
	}

	(*content)[ConfigKey] = temp[ConfigTemplate]

	err = EtcdClient.Put(ConfigKey, temp[ConfigTemplate])

	if err != nil {
		log.Fatal("Put config faild:", "fail:", err)
	}

	log.Printf("NewConfigInEtcd:%+v", *content)
}
