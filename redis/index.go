package redis

import (
	"encoding/json"

	"github.com/go-redis/redis"
)

func UnmarshalConfig(data []byte) (Config, error) {
	var r Config
	err := json.Unmarshal(data, &r)
	return r, err
}

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
func New(c Config) *redis.Client {
	r := redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: c.Password, // no password
		DB:       int(c.DB),  // use default DB
	})

	e := r.Ping().Err()
	if e != nil {
		panic(e.Error())
	}

	if false {

		e = r.FlushAll().Err()
		if e != nil {
			panic(e.Error())
		}

	}

	return r
}
