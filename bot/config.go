package bot

import "time"

type Config struct {
	NotifyList       []*Notify
	RedisAddr        string
	RedisIdleTimeout time.Duration
	RedisMaxIdle     int
}
