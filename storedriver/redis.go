package storedriver

import (
	"github.com/go-redis/redis"
	"context"
)
var ctx = context.Background()

type RedisClient struct {
	Address string
	Passwd string
}
func (rc * RedisClient) Connect() (*redis.Client, error){
	rclient := redis.NewClient(&redis.Options{Addr: rc.Address,Password: rc.Passwd,DB: 0})
	_,err:=rclient.Ping().Result()
	return rclient,err
}
func NewRedisConfig(address string, password string) *RedisClient{
	return &RedisClient{
		Address: address,
		Passwd: password,
	}
}
