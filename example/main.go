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

func onValueChangedSubscriber(event redconf.OnValueChangedEvent) {
	var err error
	var data []byte
	if data, err = json.MarshalIndent(&event, "", "    "); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
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

	redConf.Subscribe(onValueChangedSubscriber)

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
				Clear()
				fmt.Println(currentString)
				preString = currentString
			}
		}

		time.Sleep(time.Second)
	}
}
