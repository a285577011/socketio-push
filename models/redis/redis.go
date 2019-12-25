// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package redis for cache provider
//
// depend on github.com/gomodule/redigo/redis
//
// go install github.com/gomodule/redigo/redis
//
// Usage:
// import(
//   _ "github.com/astaxie/beego/cache/redis"
//   "github.com/astaxie/beego/cache"
// )
//
//  bm, err := cache.NewCache("redis", `{"conn":"127.0.0.1:11211"}`)
//
//  more docs http://beego.me/docs/module/cache.md
package redis

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"strings"
	"github.com/astaxie/beego"
)

var (
	// DefaultKey the collection name of redis for cache adapter.
	DefaultKey = "gcs"
)

// Cache is Redis cache adapter.
type Redis struct {
	p        *redis.Pool // redis connection pool
	conninfo string
	dbNum    int
	key      string
	password string
	maxIdle  int
}
var RedisModel *Redis
// NewRedisCache create new redis cache with default collection name.
func NewRedis() *Redis {
	redis:=&Redis{key: DefaultKey,dbNum:6}
	beego.LoadAppConfig("ini", "./conf/db.conf")
	conifg := map[string]string{
		"conn":     beego.AppConfig.String("redis::ip") + ":" + beego.AppConfig.String("redis::port"),
		"password": beego.AppConfig.String("redis::passw"),
		"maxIdle":beego.AppConfig.String("redis::maxIdle"),
	}
	err := redis.StartAndGC(conifg)
	if err != nil {
		panic(err)
		redis = nil
	}
	return redis
}
// actually do the redis cmds, args[0] must be the key name.
func (rc *Redis) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	if len(args) < 1 {
		return nil, errors.New("missing required arguments")
	}
	args[0] = rc.associate(args[0])
	c := rc.p.Get()
	defer c.Close()

	return c.Do(commandName, args...)
}

// associate with config key.
func (rc *Redis) associate(originKey interface{}) string {
	return fmt.Sprintf("%s:%s", rc.key, originKey)
}

// Get cache from redis.
func (rc *Redis) Get(key string) interface{} {
	if v, err := rc.do("GET", key); err == nil {
		return v
	}
	return nil
}

// GetMulti get cache from redis.
func (rc *Redis) GetMulti(keys []string) []interface{} {
	c := rc.p.Get()
	defer c.Close()
	var args []interface{}
	for _, key := range keys {
		args = append(args, rc.associate(key))
	}
	values, err := redis.Values(c.Do("MGET", args...))
	if err != nil {
		return nil
	}
	return values
}

// Put put cache to redis.
func (rc *Redis) Put(key string, val interface{}, timeout time.Duration) error {
	_, err := rc.do("SETEX", key, int64(timeout/time.Second), val)
	return err
}

// Delete delete cache in redis.
func (rc *Redis) Delete(key string) error {
	_, err := rc.do("DEL", key)
	return err
}

// IsExist check cache's existence in redis.
func (rc *Redis) IsExist(key string) bool {
	v, err := redis.Bool(rc.do("EXISTS", key))
	if err != nil {
		return false
	}
	return v
}

// Incr increase counter in redis.
func (rc *Redis) Incr(key string) (int, error) {
	incrNum, err := redis.Int(rc.do("INCRBY", key, 1))
	return incrNum, err
}

// Decr decrease counter in redis.
func (rc *Redis) Decr(key string) (int, error) {
	decrNum, err := redis.Int(rc.do("DECRBY", key, 1))
	return decrNum, err
}
// Incr increase counter in redis.
func (rc *Redis) IncrByNum(key string,num int) (int, error) {
	incrNum, err := redis.Int(rc.do("INCRBY", key, num))
	return incrNum, err
}

// Decr decrease counter in redis.
func (rc *Redis) DecrByNum(key string,num int) (int, error) {
	decrNum, err := redis.Int(rc.do("DECRBY", key, num))
	return decrNum, err
}
// Put put cache to redis.
func (rc *Redis) Set(key string, val interface{}) error {
	_, err := rc.do("SET", key, val)
	return err
}
// ClearAll clean all cache in redis. delete this redis collection.
func (rc *Redis) ClearAll() error {
	c := rc.p.Get()
	defer c.Close()
	cachedKeys, err := redis.Strings(c.Do("KEYS", rc.key+":*"))
	if err != nil {
		return err
	}
	for _, str := range cachedKeys {
		if _, err = c.Do("DEL", str); err != nil {
			return err
		}
	}
	return err
}
// Put put cache to redis.
func (rc *Redis) Hset(key string,field string, val interface{}) error {
	_, err := rc.do("HSET", key, field,val)
	return err
}
func (rc *Redis) Hdel(key string,field string) error {
	_, err := rc.do("HDEL", key, field)
	return err
}
// StartAndGC start redis cache adapter.
// config is like {"key":"collection key","conn":"connection info","dbNum":"0"}
// the cache item in redis are stored forever,
// so no gc operation.
func (rc *Redis) StartAndGC(cf map[string]string) error {
	//var cf map[string]string
	//json.Unmarshal([]byte(config), &cf)

	if _, ok := cf["key"]; !ok {
		cf["key"] = DefaultKey
	}
	if _, ok := cf["conn"]; !ok {
		return errors.New("config has no conn key")
	}

	// Format redis://<password>@<host>:<port>
	cf["conn"] = strings.Replace(cf["conn"], "redis://", "", 1)
	if i := strings.Index(cf["conn"], "@"); i > -1 {
		cf["password"] = cf["conn"][0:i]
		cf["conn"] = cf["conn"][i+1:]
	}

	//if _, ok := cf["dbNum"]; !ok {
		//cf["dbNum"] = "0"
	//}
	if _, ok := cf["password"]; !ok {
		cf["password"] = ""
	}
	if _, ok := cf["maxIdle"]; !ok {
		cf["maxIdle"] = "3"
	}
	rc.key = cf["key"]
	rc.conninfo = cf["conn"]
	//rc.dbNum, _ = strconv.Atoi(cf["dbNum"])
	rc.password = cf["password"]
	rc.maxIdle, _ = strconv.Atoi(cf["maxIdle"])

	rc.connectInit()

	c := rc.p.Get()
	defer c.Close()

	return c.Err()
}

// connect to redis.
func (rc *Redis) connectInit() {
	dialFunc := func() (c redis.Conn, err error) {
		c, err = redis.Dial("tcp", rc.conninfo)
		if err != nil {
			return nil, err
		}

		if rc.password != "" {
			if _, err := c.Do("AUTH", rc.password); err != nil {
				c.Close()
				return nil, err
			}
		}

		_, selecterr := c.Do("SELECT", rc.dbNum)
		if selecterr != nil {
			c.Close()
			return nil, selecterr
		}
		return
	}
	// initialize a new pool
	rc.p = &redis.Pool{
		MaxIdle:     rc.maxIdle,
		IdleTimeout: 180 * time.Second,
		Dial:        dialFunc,
	}
}
func init(){
	RedisModel=NewRedis()
}