package redis

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/suiguo/hwlib/logger"

	"github.com/redis/go-redis/v9"
)

var instanceMap = make(map[string]*Client)
var redis_lock sync.RWMutex

func createTLSConfiguration(cfg *TlsCfg) (t *tls.Config) {
	if cfg == nil {
		return nil
	}
	t = &tls.Config{
		InsecureSkipVerify: cfg.Skip,
	}
	if cfg.CertFile != "" && cfg.KeyFile != "" && cfg.CaFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			log.Fatal(err)
		}

		caCert, err := os.ReadFile(cfg.CaFile)
		if err != nil {
			log.Fatal(err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		t = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: cfg.Skip,
		}
	}
	return t
}

// Client is for
type Client struct {
	Cc  redis.Cmdable
	log logger.Logger
}

const ResultNil = redis.Nil

type TlsCfg struct {
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
	CaFile   string `json:"ca_file"`
	Skip     bool   `json:"skip"`
	// certFile string, keyFile string, caFile string, skip bool
}
type RedisCfg struct {
	IsCluster bool        `json:"is_cluster"`
	DbIdx     int         `json:"db_idx"`
	UserName  string      `json:"user_name"`
	Url       []RedisHost `json:"url"`
	PassWord  string      `json:"pwd"`
	TlsCfg    *TlsCfg     `json:"tls"`
}
type RedisHost struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func GetInstance(log logger.Logger, cfg *RedisCfg) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("is nil")
	}
	if log == nil {
		log = logger.NewStdLogger(2)
	}
	var tls *tls.Config
	if cfg.TlsCfg != nil {
		tls = createTLSConfiguration(cfg.TlsCfg)
	}
	if key, err := json.Marshal(cfg); err == nil {
		redis_lock.RLock()
		r := instanceMap[string(key)]
		redis_lock.RUnlock()
		if r == nil {
			var rdb redis.Cmdable
			if cfg.IsCluster {
				opt := &redis.ClusterOptions{
					Username:  cfg.UserName,
					Password:  cfg.PassWord,
					TLSConfig: tls,
				}
				for _, val := range cfg.Url {
					opt.Addrs = append(opt.Addrs, fmt.Sprintf("%s:%d", val.Host, val.Port))
				}
				rdb = redis.NewClusterClient(opt)
			} else {
				if len(cfg.Url) != 1 {
					return nil, fmt.Errorf("not cluster client need len(url)=1")
				}
				rdb = redis.NewClient(&redis.Options{
					Addr:      fmt.Sprintf("%s:%d", cfg.Url[0].Host, cfg.Url[0].Port),
					Username:  cfg.UserName,
					Password:  cfg.PassWord, // no password set
					DB:        cfg.DbIdx,    // use default DB
					TLSConfig: tls,
				})
			}
			errors := rdb.Ping(context.Background()).Err()
			if errors != nil {
				return nil, errors
			}
			r = &Client{
				Cc:  rdb,
				log: log,
			}
			redis_lock.Lock()
			instanceMap[string(key)] = r
			redis_lock.Unlock()
		}
		return r, nil
	} else {
		return nil, err
	}
}

// EXPIRE is for
func (c *Client) EXPIRE(ctx context.Context, key string, dur time.Duration) error {
	if c == nil || c.Cc == nil {
		return fmt.Errorf("redis not init ")
	}
	c.Cc.Expire(ctx, key, dur)
	return nil
}

func (c *Client) SPOP(ctx context.Context, key string) (string, error) {
	if c == nil || c.Cc == nil {
		return "", fmt.Errorf("redis not init")
	}
	return c.Cc.SPop(ctx, key).Result()
}

// SADD is for
func (c *Client) SADD(ctx context.Context, key string, values ...interface{}) error {
	if c == nil || c.Cc == nil {
		return fmt.Errorf("redis not init ")
	}
	return c.Cc.SAdd(ctx, key, values...).Err()
}

// SREM is for
func (c *Client) SREM(ctx context.Context, key string, values ...interface{}) error {
	if c == nil || c.Cc == nil {
		return fmt.Errorf("redis not init ")
	}
	return c.Cc.SRem(ctx, key, values...).Err()
}

// SMEMBERS is ofr
func (c *Client) SMEMBERS(ctx context.Context, key string) ([]string, error) {
	if c == nil || c.Cc == nil {
		return []string{}, fmt.Errorf("redis not init ")
	}
	return c.Cc.SMembers(ctx, key).Result()
}

// BLPOP is for
func (c *Client) BLPOP(ctx context.Context, key string, t time.Duration) ([]string, error) {
	if c == nil || c.Cc == nil {
		return []string{}, fmt.Errorf("redis not init ")
	}
	return c.Cc.BLPop(ctx, t, key).Result()
}

// RPUSH is for
func (c *Client) RPUSH(ctx context.Context, key string, value ...interface{}) (int64, error) {
	if c == nil || c.Cc == nil {
		return 0, fmt.Errorf("redis not init ")
	}
	return c.Cc.RPush(ctx, key, value).Result()
}

// RPUSH is for
func (c *Client) LRANGE(ctx context.Context, key string, start, stop int64) ([]string, error) {
	if c == nil || c.Cc == nil {
		return nil, fmt.Errorf("redis not init ")
	}
	return c.Cc.LRange(ctx, key, start, stop).Result()
}

// SET is for
func (c *Client) SET(ctx context.Context, key string, value interface{}, t time.Duration) (string, error) {
	if c == nil || c.Cc == nil {
		return "", fmt.Errorf("redis not init ")
	}
	return c.Cc.Set(ctx, key, value, t).Result()
}

// SETNX is for
func (c *Client) SETNX(ctx context.Context, key string, value interface{}, t time.Duration) (bool, error) {
	if c == nil || c.Cc == nil {
		return false, fmt.Errorf("redis not init ")
	}
	return c.Cc.SetNX(ctx, key, value, t).Result()
}

// DEL is for
func (c *Client) DEL(ctx context.Context, key ...string) (int64, error) {
	if c == nil || c.Cc == nil {
		return 0, fmt.Errorf("redis not init ")
	}
	return c.Cc.Del(ctx, key...).Result()
}

// GET is for
func (c *Client) GET(ctx context.Context, key string) (string, error) {
	if c == nil || c.Cc == nil {
		return "", fmt.Errorf("redis not init ")
	}
	return c.Cc.Get(ctx, key).Result()
}

// XADDJSON is for add message to stream, if stream not exist, stream will create.
// XADDJSON will format all interface{} value to json str.
func (c *Client) XADDJSON(ctx context.Context, stream string, vals map[string]interface{}) error {
	if c == nil || c.Cc == nil {
		return fmt.Errorf("redis not init ")
	}
	//reformat interface to json str
	for k, v := range vals {
		res, _ := json.Marshal(v)
		vals[k] = res
	}
	xargs := redis.XAddArgs{
		Stream: stream,
		Values: vals,
	}
	stat := c.Cc.XAdd(ctx, &xargs)
	return stat.Err()
}

// XADD interface{} value is only support string and numeric value.
func (c *Client) XADD(ctx context.Context, stream, id string, maxlen int64, vals map[string]interface{}) error {
	if c == nil || c.Cc == nil {
		return fmt.Errorf("redis not init ")
	}
	xargs := redis.XAddArgs{
		Stream: stream,
		ID:     id,
		MaxLen: maxlen,
		Values: vals,
	}
	stat := c.Cc.XAdd(ctx, &xargs)
	return stat.Err()
}

// XGROUP_CREATE create consumer group.
func (c *Client) XGROUP_CREATE(ctx context.Context, stream, group, start string) (string, error) {
	if c == nil || c.Cc == nil {
		return "", fmt.Errorf("redis not init ")
	}
	stat := c.Cc.XGroupCreate(ctx, stream, group, start)
	return stat.Result()
}

// XGROUP_DELETE delete consumer group.
func (c *Client) XGROUP_DELETE(ctx context.Context, stream, group string) error {
	if c == nil || c.Cc == nil {
		return fmt.Errorf("redis not init ")
	}
	stat := c.Cc.XGroupDestroy(ctx, stream, group)
	return stat.Err()
}

// XGROUP_SETID modify consumer group start.
func (c *Client) XGROUP_SETID(ctx context.Context, stream, group, start string) error {
	if c == nil || c.Cc == nil {
		return fmt.Errorf("redis not init ")
	}
	stat := c.Cc.XGroupSetID(ctx, stream, group, start)
	return stat.Err()
}

// XGROUP_DELCONSUMER delete consumer from group.
func (c *Client) XGROUP_DELCONSUMER(ctx context.Context, stream, group, consumer string) error {
	if c == nil || c.Cc == nil {
		return fmt.Errorf("redis not init ")
	}
	stat := c.Cc.XGroupDelConsumer(ctx, stream, group, consumer)
	return stat.Err()
}

// XGROUP_READ is for read message from stream
func (c *Client) XGROUP_READ(ctx context.Context, stream, group, consumer, start string, count int64, timeout time.Duration, noAck bool, jsonDecode bool) (*map[string]map[string]interface{}, error) {
	rgArgs := redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		// List of streams and ids.
		Streams: []string{stream, start},
		Count:   count,
		Block:   timeout,
		NoAck:   noAck,
	}
	stat := c.Cc.XReadGroup(ctx, &rgArgs)
	err := stat.Err()
	if err != nil {
		return nil, err
	}
	xstreams, err := stat.Result()
	if err != nil {
		return nil, err
	}
	result := make(map[string]map[string]interface{})
	for _, stream := range xstreams {
		for _, msg := range stream.Messages {
			ID := msg.ID
			if jsonDecode {
				for k, v := range msg.Values {
					var tmp interface{}
					err = json.Unmarshal([]byte(fmt.Sprint(v)), &tmp)
					if err != nil {
						return nil, err
					}
					msg.Values[k] = tmp
				}
			}
			values := msg.Values
			result[ID] = values
		}
	}
	return &result, nil
}

// XPENDING_SCAN is for read message only received not ack.
func (c *Client) XPENDING_SCAN(ctx context.Context, stream, group, consumer, start, end string, count int64) (*map[string]map[string]interface{}, error) {
	rgArgs := redis.XPendingExtArgs{
		Stream:   stream,
		Group:    group,
		Start:    start,
		End:      end,
		Count:    count,
		Consumer: consumer,
	}
	stat := c.Cc.XPendingExt(ctx, &rgArgs)
	err := stat.Err()
	if err != nil {
		return nil, err
	}
	xstreams, err := stat.Result()
	if err != nil {
		return nil, err
	}
	result := make(map[string]map[string]interface{})
	for _, stream := range xstreams {
		tmp := make(map[string]interface{})
		tmp["consumer"] = stream.Consumer
		tmp["idle"] = stream.Idle
		tmp["retryCount"] = stream.RetryCount
		result[stream.ID] = tmp
	}
	return &result, nil
}

// XACK ack target message.
func (c *Client) XACK(ctx context.Context, stream, group string, ids ...string) error {
	if c == nil || c.Cc == nil {
		return fmt.Errorf("redis not init ")
	}
	stat := c.Cc.XAck(ctx, stream, group, ids...)
	return stat.Err()
}

// XINFO_GROUPS is for get stream group info
func (c *Client) XINFO_GROUPS(ctx context.Context, stream string) ([]redis.XInfoGroup, error) {
	if c == nil || c.Cc == nil {
		return nil, fmt.Errorf("redis not init ")
	}
	return c.Cc.XInfoGroups(ctx, stream).Result()
}

// XCLAIM reclaim pending message to target consumer.
func (c *Client) XCLAIM(ctx context.Context, stream, group, consumer string, minIdle time.Duration, ids ...string) error {
	if c == nil || c.Cc == nil {
		return fmt.Errorf("redis not init ")
	}
	rgArgs := redis.XClaimArgs{
		Stream:   stream,
		Group:    group,
		Consumer: consumer,
		MinIdle:  minIdle,
		Messages: ids,
	}
	stat := c.Cc.XClaim(ctx, &rgArgs)
	return stat.Err()
}

// HGET is for
func (c *Client) HGET(ctx context.Context, key string, field string) (string, error) {
	if c == nil || c.Cc == nil {
		return "", fmt.Errorf("redis not init ")
	}
	return c.Cc.HGet(ctx, key, field).Result()
}

// HGETALL is for
func (c *Client) HGETALL(ctx context.Context, key string) (map[string]string, error) {
	if c == nil || c.Cc == nil {
		return map[string]string{}, fmt.Errorf("redis not init ")
	}
	return c.Cc.HGetAll(ctx, key).Result()
}

// HSET is for
func (c *Client) HSET(ctx context.Context, name string, key string, value interface{}) error {
	if c == nil || c.Cc == nil {
		return fmt.Errorf("redis not init ")
	}
	return c.Cc.HSet(ctx, name, key, value).Err()
}

// HMGET is for
func (c *Client) HMGET(ctx context.Context, name string, keys ...string) ([]interface{}, error) {
	if c == nil || c.Cc == nil {
		return nil, fmt.Errorf("redis not init ")
	}
	return c.Cc.HMGet(ctx, name, keys...).Result()
}

// HMSET is for
func (c *Client) HMSET(ctx context.Context, name string, kv map[string]interface{}) error {
	if c == nil || c.Cc == nil {
		return fmt.Errorf("redis not init ")
	}
	return c.Cc.HMSet(ctx, name, kv).Err()
}

// HGET is for
func (c *Client) HDEL(ctx context.Context, key string, field string) (int64, error) {
	if c == nil || c.Cc == nil {
		return 0, fmt.Errorf("redis not init ")
	}
	return c.Cc.HDel(ctx, key, field).Result()
}

func (c *Client) HEXISTS(ctx context.Context, key string, field string) (bool, error) {
	if c == nil || c.Cc == nil {
		return false, fmt.Errorf("redis not init ")
	}
	return c.Cc.HExists(ctx, key, field).Result()
}

func (c *Client) EXISTS(ctx context.Context, key string) (bool, error) {
	if c == nil || c.Cc == nil {
		return false, fmt.Errorf("redis not init ")
	}
	r, err := c.Cc.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return r > 0, err
}

// ZADD is for
func (c *Client) ZADD(ctx context.Context, name string, score float64, member interface{}) error {
	if c == nil || c.Cc == nil {
		return fmt.Errorf("redis not init ")
	}
	tmp := redis.Z{
		Score:  score,
		Member: member,
	}
	return c.Cc.ZAdd(ctx, name, tmp).Err()
}

func (c *Client) ZAdd(ctx context.Context, key string, members any, score float64) (int64, error) {
	if c == nil || c.Cc == nil {
		return 0, fmt.Errorf("redis not init ")
	}
	return c.Cc.ZAdd(ctx, key, redis.Z{Score: score, Member: members}).Result()
}

// ZRevRangeByScore is for
func (c *Client) ZRevRangeByScore(ctx context.Context, key, min, max string, count int64) ([]string, error) {
	if c == nil || c.Cc == nil {
		return nil, fmt.Errorf("redis not init ")
	}
	opt := &redis.ZRangeBy{
		Min:   min,
		Max:   max,
		Count: count,
	}
	return c.Cc.ZRevRangeByScore(ctx, key, opt).Result()
}

// ZRANGE is for
func (c *Client) ZRANGE(ctx context.Context, key string, start int64, stop int64) ([]string, error) {
	if c == nil || c.Cc == nil {
		return nil, fmt.Errorf("redis not init ")
	}
	return c.Cc.ZRange(ctx, key, start, stop).Result()
}

// ZREM is for
func (c *Client) ZREM(ctx context.Context, name string, member string) error {
	if c == nil || c.Cc == nil {
		return fmt.Errorf("redis not init ")
	}
	return c.Cc.ZRem(ctx, name, member).Err()
}

// ZPOPMIN is for
func (c *Client) ZPOPMIN(ctx context.Context, key string) ([]redis.Z, error) {
	if c == nil || c.Cc == nil {
		return nil, fmt.Errorf("redis not init ")
	}
	return c.Cc.ZPopMin(ctx, key).Result()
}

// ZPOPMAX is for
func (c *Client) ZPOPMAX(ctx context.Context, key string) ([]redis.Z, error) {
	if c == nil || c.Cc == nil {
		return nil, fmt.Errorf("redis not init ")
	}
	return c.Cc.ZPopMax(ctx, key).Result()
}

func (c *Client) ScriptLoad(ctx context.Context, script string) (string, error) {
	if c == nil || c.Cc == nil {
		return "", fmt.Errorf("redis not init ")
	}
	return c.Cc.ScriptLoad(ctx, script).Result()
}
func (c *Client) ScriptExists(ctx context.Context, sha ...string) ([]bool, error) {
	if c == nil || c.Cc == nil {
		return nil, fmt.Errorf("redis not init ")
	}
	return c.Cc.ScriptExists(ctx, sha...).Result()
}

func (c *Client) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	if c == nil || c.Cc == nil {
		return "", fmt.Errorf("redis not init ")
	}
	return c.Cc.EvalSha(ctx, sha1, keys, args...).Result()
}

func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	if c == nil || c.Cc == nil {
		return "", fmt.Errorf("redis not init ")
	}
	return c.Cc.Eval(ctx, script, keys, args...).Result()
}

// INCR is for
func (c *Client) INCR(ctx context.Context, key string, value int64) (int64, error) {
	if c == nil || c.Cc == nil {
		return 0, fmt.Errorf("redis not init ")
	}
	return c.Cc.IncrBy(ctx, key, value).Result()
}

func (c *Client) HINCR(ctx context.Context, key string, feild string, value int64) (int64, error) {
	if c == nil || c.Cc == nil {
		return 0, fmt.Errorf("redis not init ")
	}
	if r := c.Cc.HIncrBy(ctx, key, feild, value); r != nil {
		return r.Result()
	}

	return 0, fmt.Errorf("result is nil")
}

func (c *Client) HINCRFLOAT(ctx context.Context, key string, feild string, value float64) (float64, error) {
	if c == nil || c.Cc == nil {
		return 0, fmt.Errorf("redis not init ")
	}
	if r := c.Cc.HIncrByFloat(ctx, key, feild, value); r != nil {
		return r.Result()
	}

	return 0, fmt.Errorf("result is nil")
}

func (c *Client) HKEYS(ctx context.Context, key string) ([]string, error) {
	if c == nil || c.Cc == nil {
		return nil, fmt.Errorf("redis not init ")
	}
	if r := c.Cc.HKeys(ctx, key); r != nil {
		return r.Result()
	}
	return nil, fmt.Errorf("result is nil")
}

func (c *Client) KEYS(ctx context.Context, key string) ([]string, error) {
	if c == nil || c.Cc == nil {
		return nil, fmt.Errorf("redis not init ")
	}
	if r := c.Cc.Keys(ctx, key); r != nil {
		return r.Result()
	}
	return nil, fmt.Errorf("result is nil")
}

// HIncr

func (c *Client) Pipeline() (redis.Pipeliner, error) {
	if c == nil || c.Cc == nil {
		return nil, fmt.Errorf("redis not init ")
	}
	return c.Cc.Pipeline(), nil
}
