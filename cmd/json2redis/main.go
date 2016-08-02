package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	app := cli.NewApp()

	app.HideVersion = true

	app.Action = sync

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "redis-host,host",
			Usage: "Redis host",
			Value: "localhost",
		},
		cli.IntFlag{
			Name:  "redis-port,port",
			Usage: "Redis port",
			Value: 6379,
		},
		cli.StringFlag{
			Name:  "redis-password",
			Usage: "Redis password",
			Value: "",
		},
		cli.IntFlag{
			Name:  "redis-db,db",
			Usage: "Redis database index",
			Value: 6379,
		},
		cli.StringFlag{
			Name:  "namespace,n",
			Usage: "Key's namespace",
		},
		cli.StringFlag{
			Name:  "channel",
			Usage: "Which redis channel to publish value changed event",
			Value: "REDCONF:ONCHANGED",
		},
		cli.BoolFlag{
			Name:  "notify",
			Usage: "Publish changed event to redis-channel",
		},
		cli.StringFlag{
			Name:  "filename,f",
			Usage: "JSON file for import to redis",
		},
		cli.StringFlag{
			Name:  "config-name",
			Usage: "name of config struct, defualt will use josn filename(exclude file ext)",
		},
		cli.StringFlag{
			Name:  "workdir,w",
			Usage: "change work dir before sync",
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func sync(ctx *cli.Context) (err error) {
	workdir := ctx.String("workdir")
	if len(workdir) > 0 {
		if err = os.Chdir(workdir); err != nil {
			return
		}
	}

	filename := ctx.String("filename")
	if len(filename) == 0 {
		err = fmt.Errorf("JSON filename not specfic")
		return
	}

	configName := ctx.String("config-name")

	if len(configName) == 0 {
		configName = filepath.Base(filename)
		ext := filepath.Ext(configName)

		configName = strings.TrimRight(configName, ext)
	}

	channel := ctx.String("channel")
	notify := ctx.Bool("notify")
	if notify && len(channel) == 0 {
		err = fmt.Errorf("notify channel is empty")
		return
	}

	var kv map[string]string
	if kv, err = loadData(filename, configName); err != nil {
		return
	}

	host := ctx.String("redis-host")
	port := ctx.Int("redis-port")
	db := ctx.Int("redis-db")
	pwd := ctx.String("redis-password")
	namespace := ctx.String("namespace")

	changed, errs := pushToRedis(host, port, pwd, db, namespace, kv, notify, channel)

	if errs != nil {
		fmt.Println("ERRORS:\n-----------------------------------")
		fmt.Println(errs.Error(), "\n")
	}

	if len(changed) > 0 {
		fmt.Println("CHANGES:\n----------------------------------")
		for k, v := range changed {
			fmt.Printf("%s: %s\n", k, v)
		}
		fmt.Println("")
	}

	fmt.Printf("%d key changed and syned\n", len(changed))

	return
}

func pushToRedis(host string, port int, password string, db int, namespace string, data map[string]string, notify bool, channel string) (changed map[string]string, errs error) {

	if port == 0 {
		port = 6379
	}

	address := fmt.Sprintf("%s:%d", host, port)

	pool := redis.NewPool(
		func() (conn redis.Conn, e error) {
			conn, e = redis.Dial("tcp", address)
			if e != nil {
				return
			}

			if len(password) > 0 {
				if _, e = conn.Do("AUTH", password); e != nil {
					conn.Close()
					return
				}
			}

			if _, e = conn.Do("SELECT", db); e != nil {
				return
			}
			return
		}, 0)

	syncFailure := map[string]string{}
	notifyFailure := map[string]string{}

	changed = map[string]string{}

	for k, v := range data {
		if len(k) == 0 {
			continue
		}

		key := k
		if len(namespace) > 0 {
			key = namespace + ":" + k
		}

		conn := pool.Get()
		if ret, e := conn.Do("GET", key); e != nil {
			syncFailure[key] = e.Error()
			continue
		} else {
			conn = pool.Get()
			oldV, _ := redis.String(ret, e)

			if oldV != v {
				if _, e := conn.Do("SET", key, v); e != nil {
					syncFailure[key] = e.Error()
					continue
				}

				changed[key] = fmt.Sprintf("%s ==> %s", oldV, v)

				if notify {
					conn = pool.Get()
					if _, e := conn.Do("PUBLISH", channel, key); e != nil {
						notifyFailure[key] = e.Error()
					}
				}
			}
		}
	}

	strBuf := ""

	if len(syncFailure) > 0 {
		for k, v := range syncFailure {
			strBuf += fmt.Sprintf("SYNC_ERROR: %s: %s\n", k, v)
		}
	}

	if len(notifyFailure) > 0 {
		for k, v := range notifyFailure {
			strBuf += fmt.Sprintf("NOTIFY_ERROR: %s: %s\n", k, v)
		}
	}

	if len(strBuf) > 0 {
		errs = fmt.Errorf(strBuf)
	}

	return
}

func loadData(filename, configName string) (kv map[string]string, err error) {
	var data []byte
	if data, err = ioutil.ReadFile(filename); err != nil {
		return
	}

	buf := bytes.NewBuffer(data)
	decoder := json.NewDecoder(buf)
	decoder.UseNumber()

	tmpMap := map[string]interface{}{}

	if err = decoder.Decode(&tmpMap); err != nil {
		return
	}

	if tmpMap == nil {
		return
	}

	resultKV := map[string]string{}

	deepInMap(&configName, tmpMap, resultKV)

	kv = resultKV
	return
}

func deepInMap(prefix *string, m interface{}, resultKV map[string]string) {
	switch typedM := m.(type) {
	case map[string]interface{}:
		{
			for k, v := range typedM {
				newPrefix := *prefix
				if len(newPrefix) > 0 {
					newPrefix += ":" + k
				} else {
					newPrefix = k
				}
				deepInMap(&newPrefix, v, resultKV)
			}
		}
	default:
		if m != nil {
			switch v := m.(type) {
			case []interface{}:
				{
					var tmpStrV []string
					for i := 0; i < len(v); i++ {
						tmpStrV = append(tmpStrV, fmt.Sprintf("%v", v[i]))
					}
					resultKV[*prefix] = strings.Join(tmpStrV, ",")
				}
			default:
				{
					resultKV[*prefix] = fmt.Sprintf("%v", v)
				}
			}
		} else {
			resultKV[*prefix] = ""
		}
	}
}
