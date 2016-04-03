package main

import (
	"encoding/json"
	"fmt"
	"github.com/gogap/redconf"
	"strings"
	"time"
)

type Account struct {
	Name     string
	Password string
}

type ServerConfig struct {
	Host     *string
	Port     int
	AllowIPs []string
}

type LogConfig struct {
	Path    string
	Maxsize int
}

type AppConfig struct {
	Server    ServerConfig
	Log       LogConfig
	Accounts  []Account
	BlackList []string
}

func main() {

	var err error
	var monitor redconf.Monitor
	var storage redconf.Storage
	var opts redconf.Options

	channel := "ONCHANGED"

	opts = redconf.Options{
		"address":  "localhost:6379",
		"password": "",
		"db":       0,
		"idle":     10,
		"channel":  channel,
	}

	if monitor, err = redconf.CreateMonitor("redis", opts); err != nil {
		fmt.Println(err)
		return
	}

	if storage, err = redconf.CreateStorage("redis", opts); err != nil {
		fmt.Println(err)
		return
	}

	var redConf *redconf.RedConf
	namespace := "GOGAP"

	if redConf, err = redconf.New(namespace, storage, monitor); err != nil {
		return
	}

	appConf := AppConfig{}

	if err = redConf.Watch(&appConf); err != nil {
		fmt.Println(err)
		return
	}

	keys := redConf.Keys()

	strKeys := strings.Join(keys, ",\n")

	preString := ""

	for {
		var data []byte
		if data, err = json.MarshalIndent(&appConf, "", "    "); err != nil {
			fmt.Println(err)
			return
		} else {

			currentString := fmt.Sprintf(
				"namespace:\t%s\nchannel:\t%s\n==========================\n%s\n==========================\n%s",
				namespace,
				channel,
				strKeys,
				string(data))

			if currentString != preString {
				//Clear()
				fmt.Println(currentString)
				preString = currentString
			}

			fmt.Printf("\r%s", time.Now().Format("2006-01-02 15:04:05"))
		}

		time.Sleep(time.Second)
	}
}
