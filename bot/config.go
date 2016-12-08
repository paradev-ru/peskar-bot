package bot

import "time"

type Config struct {
	Actions          []*Action
	RedisAddr        string
	RedisIdleTimeout time.Duration
	RedisMaxIdle     int
}
