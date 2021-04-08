package redis

import (
	"github.com/go-redis/redis"
)

type IConfig interface {
	Host() string
	Password() string
	Index() int64
}

//-------------------------------------------------------------------------------------------------

// New ...
func New(c IConfig) *redis.Client {
	r := redis.NewClient(&redis.Options{
		Addr:     c.Host(),
		Password: c.Password(),   // no password
		DB:       int(c.Index()), // use default DB
	})
	e := r.Ping().Err()
	if e != nil {
		panic(e.Error())
	}

	e = r.FlushAll().Err()
	if e != nil {
		panic(e.Error())
	}

	return r
}
