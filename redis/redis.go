package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/adamesong/go-util/color"
	"github.com/adamesong/go-util/logging"

	"github.com/redis/go-redis/v9"
	// "github.com/go-redis/redis"
)

// go-redis的使用：https://www.jianshu.com/p/4045a3721b3c https://segmentfault.com/a/1190000007078961
// redis 用法详解：https://www.jianshu.com/p/2639549bedc8

type RedisClient struct {
	Addr     string // 例如: xxxxx.a2fdoa.0001.usw2.cache.amazonaws.com:6379
	Password string
	DB       int // 例如：0
}

var ctx = context.Background()

func (r *RedisClient) Connect() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     r.Addr,
		Password: r.Password,
		DB:       r.DB,
	})

	//pong, err := client.Ping().Result()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Println(color.Red("Redis连接测试失败！！！"))
		fmt.Println(err.Error())
		logging.Error(err.Error())
		//panic("Redis连接失败！！")
	}
	//else {
	//	fmt.Printf(color.Blue("Redis连接测试成功！ %s\n"), pong)
	//}
	return client
}

func closeConnect(client *redis.Client) {
	if err := client.Close(); err != nil {
		fmt.Println(color.Red("redis client 关闭失败"))
		logging.Error("redis client 关闭失败")
	}
	//else {
	//	fmt.Println(color.Blue("redis client 关闭成功"))
	//}
}

// Keys 用于查找所有符合给定模式 pattern 的 key
func (r *RedisClient) Keys(pattern string) (keys []string, err error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.Keys(ctx, pattern).Result()
}

// Set 设置key value，如果duration为0，则意味着无有效期，永远存在
func (r *RedisClient) Set(key string, value interface{}, duration time.Duration) error {
	client := r.Connect()
	defer closeConnect(client)
	err := client.Set(ctx, key, value, duration).Err()
	return err
}

// Get 返回[]byte，如果不存在，则error为 redis: nil
func (r *RedisClient) Get(key string) ([]byte, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.Get(ctx, key).Bytes()
}

func (r *RedisClient) MGet(keys ...string) ([]interface{}, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.MGet(ctx, keys...).Result()
}

// SetNX 只有key不存在时，当前set操作才执行
func (r *RedisClient) SetNX(key string, value interface{}, duration time.Duration) (bool, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.SetNX(ctx, key, value, duration).Result()
}

// SetXX 只有key存在时，当前set操作才执行
func (r *RedisClient) SetXX(key string, value interface{}, duration time.Duration) (bool, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.SetXX(ctx, key, value, duration).Result()
}

// Exists 检查某一个或多个key是否存在，如不存在，返回的数字是0
func (r *RedisClient) Exists(key ...string) (int64, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.Exists(ctx, key...).Result()
}

// GetSet 设置新值并获取原来的值，不改变duration。注意，如果没有原来的值，则duration会是永久的
// 获取：GetSet（原子性），设置新值，返回旧值。比如一个按小时计算的计数器，可以用GetSet获取计数并重置为0。
func (r *RedisClient) GetSet(key string, value interface{}) ([]byte, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.GetSet(ctx, key, value).Bytes()
}

// Delete 返回删除的key的数量
func (r *RedisClient) Delete(keys ...string) (int64, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.Del(ctx, keys...).Result()
}

// TTL 返回 以毫秒为单位的整数值TTL或负值
// -1ns, 如果key没有到期超时（没有设置有效期）。-2ns, 如果键不存在(或者键已经过期了)。
func (r *RedisClient) TTL(key string) (time.Duration, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.TTL(ctx, key).Result()
}

// 重设timeout时间
// 如果key不存在，则返回false；如果原来没有有效期，现在会有有效期
func (r *RedisClient) Expire(key string, duration time.Duration) (bool, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.Expire(ctx, key, duration).Result()
}

func (r *RedisClient) LikeDeletes(key string) (int64, error) {
	client := r.Connect()
	defer closeConnect(client)
	keys, err := client.Keys(ctx, "*"+key+"*").Result()
	if err != nil {
		return int64(0), err
	}
	num, err := r.Delete(keys...)
	return num, err
}

// 一次放入多个value进一个list的尾部
func (r *RedisClient) RPush(key string, values ...interface{}) (int64, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.RPush(ctx, key, values...).Result()
}

// 一次放入1个value的尾部
func (r *RedisClient) RPushX(key string, value interface{}) (int64, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.RPushX(ctx, key, value).Result()
}

type Z = redis.Z // 类型别名

func (r *RedisClient) ZAdd(key string, members ...Z) (int64, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.ZAdd(ctx, key, members...).Result()
}

// 注：第一个的index是0
func (r *RedisClient) ZRange(key string, start, stop int64) ([]string, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.ZRange(ctx, key, start, stop).Result()
}

func (r *RedisClient) ZRevRange(key string, start, stop int64) ([]string, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.ZRevRange(ctx, key, start, stop).Result()
}

func (r *RedisClient) HMSet(key string, fields map[string]interface{}) (bool, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.HMSet(ctx, key, fields).Result()
}

func (r *RedisClient) MSet(pairs ...interface{}) (string, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.MSet(ctx, pairs...).Result()
}

func (r *RedisClient) MSetNX(pairs ...interface{}) (bool, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.MSetNX(ctx, pairs...).Result()
}

// 自定义的方法，通过pipeline实现批量给key设置过期时间
// https://www.jianshu.com/p/4045a3721b3c
// http://vearne.cc/archives/1113
func (r *RedisClient) MExpire(keys []string, duration time.Duration) error {
	client := r.Connect()
	defer closeConnect(client)
	pl := client.Pipeline()
	for _, key := range keys {
		pl.Expire(ctx, key, duration)
	}
	_, err := pl.Exec(ctx)
	return err
}

// HyperLogLog
// http://remcarpediem.net/2019/06/16/用户日活月活怎么统计-Redis-HyperLogLog-详解/
func (r *RedisClient) PFAdd(key string, els ...interface{}) (int64, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.PFAdd(ctx, key, els...).Result()
}

// 会将每个key的value merge去重后统计数量
func (r *RedisClient) PFCount(keys ...string) (int64, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.PFCount(ctx, keys...).Result()
}

// 自定义的方法，通过pipeline实现批量PFCount
// 返回值{"key1": 4, "key2": 18, ...}
func (r *RedisClient) MPFCount(keys []string) (resultMap map[string]int64, err error) {
	resultMap = make(map[string]int64)
	client := r.Connect()
	defer closeConnect(client)
	pl := client.Pipeline()
	for _, key := range keys {
		pl.PFCount(ctx, key)
	}
	results, err := pl.Exec(ctx) // results类型是[]redis.Cmder
	if err != nil {
		return
	}
	for _, result := range results {
		n, err := result.(*redis.IntCmd).Result() // n 就是这个key的PFCount的结果
		if err != nil {
			logging.Error(err.Error())
		} else {
			key := result.Args()[1].(string) // 这是key。result.Args() == [pfcount, key]
			resultMap[key] = n
		}
	}
	return
}

// 如果key不存在，将key设为"1"；如果key存在，key增加1
func (r *RedisClient) Incr(key string) (int64, error) {
	client := r.Connect()
	defer closeConnect(client)
	return client.Incr(ctx, key).Result()
}
