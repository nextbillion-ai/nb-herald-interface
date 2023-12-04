package utils

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

var Client RedisClient

type RedisClient interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (string, error)
	HGetAll(key string) (map[string]string, error)
	HSet(key, field string, value interface{}) error
	HDel(key string, fields ...string) error
	ZAdd(key string, members ...redis.Z) error
	ZRemRangeByScore(key, min, max string) error
	ZRangeByScore(key string, opt redis.ZRangeBy) ([]string, error)
	Zrange(key string, start, stop int64) ([]string, error)
	ZRem(key string, members interface{}) (int64, error)
	Del(key string) (int64, error)
	SetNX(key string, value interface{}, expiration time.Duration) (bool, error)
}

//go:generate mockery --name=RedisClient
type redisClientImpl struct {
	c  *redis.Client
	fc *redis.Client
}

func (c *redisClientImpl) getPrimaryClient() *redis.Client {
	if c.fc != nil {
		return c.fc
	}

	return c.c
}

func (c *redisClientImpl) getAllClients() []*redis.Client {
	var cs []*redis.Client
	if c.fc != nil {
		cs = append(cs, c.fc)
	}
	if c.c != nil {
		cs = append(cs, c.c)
	}

	return cs
}

func (c *redisClientImpl) Set(key string, value interface{}, expiration time.Duration) error {
	logrus.Debugf("redis client setting k: %s v: %#v", key, value)
	return c.getPrimaryClient().Set(key, value, expiration).Err()
}

func (c *redisClientImpl) Get(key string) (string, error) {
	var (
		v string
		e error = redis.Nil
	)
	for _, cl := range c.getAllClients() {
		v, e = cl.Get(key).Result()
		if e == nil {
			return v, e
		}
	}

	return v, e
}

func (c *redisClientImpl) HGetAll(key string) (map[string]string, error) {
	var (
		v map[string]string
		e error = redis.Nil
	)
	for _, cl := range c.getAllClients() {
		v, e = cl.HGetAll(key).Result()
		if e == nil {
			return v, e
		}
	}

	return v, e
}

func (c *redisClientImpl) HSet(key, field string, value interface{}) error {
	return c.getPrimaryClient().HSet(key, field, value).Err()
}

func (c *redisClientImpl) HDel(key string, fields ...string) error {
	return c.getPrimaryClient().HDel(key, fields...).Err()
}

func (c *redisClientImpl) ZAdd(key string, members ...redis.Z) error {
	return c.getPrimaryClient().ZAdd(key, members...).Err()
}

func (c *redisClientImpl) ZRemRangeByScore(key, min, max string) error {
	return c.getPrimaryClient().ZRemRangeByScore(key, min, max).Err()
}

func (c *redisClientImpl) ZRangeByScore(key string, opt redis.ZRangeBy) ([]string, error) {
	var (
		v []string
		e error = redis.Nil
	)
	for _, cl := range c.getAllClients() {
		v, e = cl.ZRangeByScore(key, opt).Result()
		if e == nil {
			return v, e
		}
	}

	return v, e
}

func (c *redisClientImpl) Zrange(key string, start, stop int64) ([]string, error) {
	var (
		v []string
		e error = redis.Nil
	)
	for _, cl := range c.getAllClients() {
		v, e = cl.ZRange(key, start, stop).Result()
		if e == nil {
			return v, e
		}
	}

	return v, e
}

func (c *redisClientImpl) ZRem(key string, members interface{}) (int64, error) {
	return c.getPrimaryClient().ZRem(key, members).Result()
}

func (c *redisClientImpl) Del(key string) (int64, error) {
	return c.getPrimaryClient().Del(key).Result()
}

func (c *redisClientImpl) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.getPrimaryClient().SetNX(key, value, expiration).Result()
}

// func Init(host string, rfConf *RedisFailOverConf) {
// 	var singleClient *redis.Client
// 	if host != "" {
// 		singleClient = newClient(host)
// 	}

// 	var failOverClient *redis.Client
// 	if rfConf != nil {
// 		failOverClient = newFailOverClient(rfConf)
// 	}

// 	if singleClient == nil && failOverClient == nil {
// 		panic("fail to init either single client or fail over client")
// 	}

// 	Client = &redisClientImpl{
// 		c:  singleClient,
// 		fc: failOverClient,
// 	}

// 	logrus.Infof("redis host initialised with %v", host)
// }

func newClient(host string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     host + ":6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := rdb.Ping().Result()
	if err != nil {
		panic("redis init fail")
	}
	logrus.Infof("redis newlcient initialised with %v %v", pong, err)

	return rdb
}

func newFailOverClient(conf *RedisFailOverConf) *redis.Client {
	rdb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    conf.MasterName,
		SentinelAddrs: conf.GetSentinelAddress(),
	})

	pong, err := rdb.Ping().Result()
	if err != nil {
		panic("redis failover init fail")
	}
	logrus.Infof("redis failover newlcient initialised with %v %v", pong, err)

	return rdb
}
