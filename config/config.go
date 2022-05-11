package config

import "time"

type Config struct {
	Period           time.Duration `config:"period"`
	MasterEventPub string        `config:"master_event_pub"`
	tagBlackList []string        `config:"tag_black_list"`
}

var DefaultConfig = Config{
	Period:           1 * time.Second,
	MasterEventPub: "/var/run/salt/master/master_event_pub.ipc",
	tagBlackList: make([]string, 0),
}
