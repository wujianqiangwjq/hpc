package storedriver

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis"
)

var ctx = context.Background()

type RedisClient struct {
	Address string `yaml:"address"`
	Port    string `yaml:"port"`
	Passwd  string `yaml:"password"`
}

func (rc *RedisClient) Connect() (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%s", rc.Address, rc.Port)
	rclient := redis.NewClient(&redis.Options{Addr: addr, Password: rc.Passwd, DB: 0})
	_, err := rclient.Ping(ctx).Result()
	return rclient, err
}
func NewRedisConfig(address string, password string, port string) *RedisClient {
	return &RedisClient{
		Address: address,
		Passwd:  password,
		Port:    port,
	}
}

func DeleteKeys(client *redis.Client, partin string) error {
	var er error = nil
	keys, kerr := client.Keys(ctx, partin).Result()
	if kerr == nil {
		_, er = client.Del(ctx, keys...).Result()
	}
	return er
}
func SetData(client *redis.Client, key string, data map[string]interface{}) error {
	err := client.HMSet(ctx, key, data).Err()
	return err
}
func CheckActive(client *redis.Client) bool {
	return client.Ping(ctx).Err() == nil
}

func GetKey(client *redis.Client, key string, filed []string) ([]interface{}, error) {
	var er error = errors.New("can't get value")
	val, err := client.HMGet(ctx, key, filed...).Result()
	if err != nil {
		return []interface{}{}, err
	}
	if len(val) > 0 {
		return val, err
	}

	return []interface{}{}, er
}
