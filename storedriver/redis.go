package storedriver

import (
	"fmt"
	"github.com/go-redis/redis"
	"context"
)
var ctx = context.Background()

type RedisClient struct {
	Address string `yaml:"address"`
	Port string   `yaml:"port"`
	Passwd string `yaml:"password"`
}
func (rc * RedisClient) Connect() (*redis.Client, error){
	addr := fmt.Sprintf("%s:%s",rc.Address,rc.Port)
	rclient := redis.NewClient(&redis.Options{Addr: addr,Password: rc.Passwd,DB: 0})
	_,err:=rclient.Ping().Result()
	return rclient,err
}
func NewRedisConfig(address string, password string, port string) *RedisClient{
	return &RedisClient{
		Address: address,
		Passwd: password,
		Port: port,
	}
}

func  DeleteKeys(client * redis.Client, partin string)error{
	keys,_:=client.Keys(partin).Result()
	_,delerr:=client.Del(keys...).Result()
	return delerr
}

