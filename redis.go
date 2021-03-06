package golanglibs

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisStruct struct {
	ctx                    context.Context
	Rdb                    *redis.Client
	networkErrorRetryTimes int
}

type RedisConfig struct {
	NetworkErrorRetryTimes int
	Database               int
	Password               string
}

// 用来过滤报错的信息， 如果包含有如下的某一个， 就判断为是网络错误
var redisNetworkErrorStrings = []string{
	"timeout",
	"connection reset by peer",
	"connection refused",
}

func getRedis(host string, port int, cfg ...RedisConfig) *RedisStruct {
	rop := &redis.Options{
		Addr: host + ":" + Str(port),
	}

	if len(cfg) != 0 {
		rop.Password = cfg[0].Password
		rop.DB = cfg[0].Database
	}

	rdb := redis.NewClient(rop)
	r := &RedisStruct{ctx: context.Background(), Rdb: rdb}
	r.Ping()

	if len(cfg) != 0 {
		r.networkErrorRetryTimes = cfg[0].NetworkErrorRetryTimes
	}

	return r
}

func (m *RedisStruct) Ping() string {
	pong, err := m.Rdb.Ping(m.ctx).Result()
	Panicerr(err)
	return pong
}

func (m *RedisStruct) Del(key string) {
	errortimes := 0
	var err error
	for {
		err = m.Rdb.Del(m.ctx, key).Err()
		if err != nil {
			if func(errfilter []string, errmsg string) bool {
				for _, err := range errfilter {
					if String(err).In(errmsg) {
						return true
					}
				}
				return false
			}(redisNetworkErrorStrings, err.Error()) && errortimes < m.networkErrorRetryTimes {
				errortimes += 1
				sleep(3)
			} else {
				Panicerr(err)
			}
		} else {
			break
		}
	}
}

func (m *RedisStruct) Set(key string, value string, ttl ...interface{}) {
	var t time.Duration

	if len(ttl) != 0 {
		if Typeof(ttl[0]) == "float64" {
			tt := ttl[0].(float64) * 1000
			t = time.Duration(tt) * time.Millisecond
		}
		if Typeof(ttl[0]) == "int" {
			tt := ttl[0].(int)
			t = time.Duration(tt) * time.Second
		}
	}

	errortimes := 0
	var err error
	for {
		err = m.Rdb.Set(m.ctx, key, value, t).Err()
		if err != nil {
			if func(errfilter []string, errmsg string) bool {
				for _, err := range errfilter {
					if String(err).In(errmsg) {
						return true
					}
				}
				return false
			}(redisNetworkErrorStrings, err.Error()) && errortimes < m.networkErrorRetryTimes {
				errortimes += 1
				sleep(3)
			} else {
				Panicerr(err)
			}
		} else {
			break
		}
	}
}

func (m *RedisStruct) Get(key string) *string {
	errortimes := 0
	var val string
	var err error
	for {
		val, err = m.Rdb.Get(m.ctx, key).Result()
		if err != nil && err != redis.Nil {
			if func(errfilter []string, errmsg string) bool {
				for _, err := range errfilter {
					if String(err).In(errmsg) {
						return true
					}
				}
				return false
			}(redisNetworkErrorStrings, err.Error()) && errortimes < m.networkErrorRetryTimes {
				errortimes += 1
				sleep(3)
			} else {
				Panicerr(err)
			}
		} else {
			break
		}
	}

	if err == redis.Nil {
		return nil
	}

	return &val
}

type RedisLockStruct struct {
	redis      *RedisStruct
	key        string
	timeoutsec int
}

var redisLockMutex sync.Mutex

func (m *RedisStruct) GetLock(key string, timeoutsec int) *RedisLockStruct {
	return &RedisLockStruct{
		redis:      m,
		key:        key,
		timeoutsec: timeoutsec, // 锁的超时时间, 为了防止进程崩溃而没有释放锁, 不是获取锁的超时时间
	}
}

func (m *RedisLockStruct) Acquire() {
	redisLockMutex.Lock()
	defer redisLockMutex.Unlock()

	for {
		b, err := m.redis.Rdb.SetNX(m.redis.ctx, m.key, 1, getTimeDuration(m.timeoutsec)).Result()
		Panicerr(err)
		if b {
			return
		} else {
			sleep(0.2)
		}
	}
}

func (m *RedisLockStruct) Release() {
	_, err := m.redis.Rdb.Del(m.redis.ctx, m.key).Result()
	Panicerr(err)
}
