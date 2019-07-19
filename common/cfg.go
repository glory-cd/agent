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
	LogLevel   string `json:"loglevel"`   // 日志级别
	Filename   string `json:"filename"`   // 日志文件路径
	MaxSize    int    `json:"maxsize"`    // 每个日志文件保存的最大尺寸 单位：M
	MaxBackups int    `json:"maxbackups"` // 日志文件最多保存多少个备份
	MaxAge     int    `json:"maxage"`     // 文件最多保存多少天
	Compress   bool   `json:"compress"`   // 是否压缩
}

type Rest struct {
	Addr string `json:"addr"`
}

type GlobalConfig struct {
	Debug bool         `json:"debug"`
	Redis *RedisConfig `json:"redis"`
	Rest  *Rest        `json:"rest"`
	Log   *LogConfig   `json:"log"`
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

//从etcd中服务配置并初始化
func InitConfig(){

	configContent, err := EtcdClient.Get(ConfigKey, false)

	log.Printf("ConfigKey:%+v", configContent)

	if err != nil{
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

//在etcd中创建新的配置
func NewConfigInEtcd(content *map[string]string){
	temp, err := EtcdClient.Get(ConfigTemplate, false)

	if err != nil{
		log.Fatal("get config template faild:", "fail:", err)
	}

	(*content)[ConfigKey] = temp[ConfigTemplate]

	err = EtcdClient.Put(ConfigKey, temp[ConfigTemplate])

	if err != nil{
		log.Fatal("Put config faild:", "fail:", err)
	}

	log.Printf("NewConfigInEtcd:%+v", *content)
}