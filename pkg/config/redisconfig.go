package Config

import "time"

type RedisConfig struct {
	RedisAddr string

	Timeout time.Duration
}