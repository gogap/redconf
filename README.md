RedConf
=======
Sync config from redis or others storages while the key's value changed

#### Usage

- The following struct is the config what we want sync with storage while the values changed

```go
type ServerConfig struct {
	Host     string
	Port     int
	AllowIPs []string
}

type LogConfig struct {
	Path    string
	Maxsize int
}

type AppConfig struct {
	Server ServerConfig
	Log    LogConfig
}
```

- We need create storage for tell redconf where the config values stored, and create monitor to notify the redconf while the values changed

```go

	opts = redconf.Options{
		"address":  "localhost:6379",
		"password": "",
		"db":       0,
		"idle":     10,
		"channel":  "ONCHANGED",
	}

	if monitor, err = redconf.CreateMonitor("redis", opts); err != nil {
		fmt.Println(err)
		return
	}

	if storage, err = redconf.CreateStorage("redis", opts); err != nil {
		fmt.Println(err)
		return
	}
```

- Create RedConf instance and watch the config

```go
	if redConf, err = redconf.New(namespace, storage, monitor); err != nil {
		return
	}

	appConf := AppConfig{}

	if err = redConf.Watch(&appConf); err != nil {
		fmt.Println(err)
		return
	}
```

- Initial redis key-value

```bash
$> redis-cli
127.0.0.1:6379> SET GOGAP:AppConfig:Server:AllowIPs 127.0.0.1,202.10.5.123
OK

```

- Run example code

``` bash
$> go run example/*.go
```

- Open new terminal session and change the config in redis

```bash
$> redis-cli
127.0.0.1:6379>SET GOGAP:AppConfig:Server:AllowIPs 127.0.0.1,202.10.5.125
OK
127.0.0.1:6379> PUBLISH ONCHANGED GOGAP:AppConfig:Server:AllowIPs
(integer) 1
```

Then you will see the change from your terminal