package redis

import (
	"github.com/go-redis/redis"
)

var Method *redis.Client

//-------------------------------------------------------------------------------------------------

type Config struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int64  `json:"db"`
}

//-------------------------------------------------------------------------------------------------

/*
Empty
是否為空值錯誤

	e error
	return true:空值錯誤, false:非空值錯誤

*/
func Empty(e error) bool {
	return e.Error() == "redis: nil"
}

// New ...
func Build(c Config) {

	Method = redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: c.Password, // no password
		DB:       int(c.DB),  // use default DB
	})

	e := Method.Ping().Err()
	if e != nil {
		panic(e.Error())
	}

	if false {

		e = Method.FlushAll().Err() // 清空 redis
		if e != nil {
			panic(e.Error())
		}

	}

}
