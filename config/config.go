package config

import "time"

type Config struct {
	Period           time.Duration `config:"period"`
	master_event_pub string        `config:"master_event_pub"`
}

var DefaultConfig = Config{
	Period:           1 * time.Second,
	master_event_pub: "/var/run/salt/master/master_event_pub.ipc",
}
