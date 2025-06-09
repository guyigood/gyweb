package redisutils

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient 是 Redis 客户端
type RedisClient struct {
	Client  *redis.Client
	Context context.Context
	IsDebug bool
}

func NewRedisClient(addr string, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx := context.Background()

	redisClient := &RedisClient{Client: client, Context: ctx}

	// 验证连接
	if err := redisClient.Ping(); err != nil {
		return nil, fmt.Errorf("Redis连接失败: %v", err)
	}
	if redisClient.IsDebug {
		fmt.Printf("[DEBUG] Redis连接成功: %s, DB: %d\n", addr, db)
	}
	return redisClient, nil
}

func (c *RedisClient) Ping() error {
	result := c.Client.Ping(c.Context)
	if err := result.Err(); err != nil {
		if c.IsDebug {
			fmt.Printf("[ERROR] Redis Ping失败: %v\n", err)
		}
		return err
	}
	if c.IsDebug {
		fmt.Printf("[DEBUG] Redis Ping成功: %s\n", result.Val())
	}
	return nil
}

// 判断redis key是否存在
func (c *RedisClient) Exists(key string) (bool, error) {
	n, err := c.Client.Exists(c.Context, key).Result()
	if err != nil {
		if c.IsDebug {
			fmt.Printf("[ERROR] Redis Exists失败 key=%s: %v\n", key, err)
		}
		return false, err
	}
	if c.IsDebug {
		fmt.Printf("[DEBUG] Redis Exists key=%s, result=%t\n", key, n > 0)
	}
	return n > 0, err
}

// 设置redis key的值
func (c *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	err := c.Client.Set(c.Context, key, value, expiration).Err()
	if err != nil {
		if c.IsDebug {
			fmt.Printf("[ERROR] Redis Set失败 key=%s, value=%v: %v\n", key, value, err)
		}
		return err
	}
	if c.IsDebug {
		fmt.Printf("[DEBUG] Redis Set成功 key=%s, value=%v, expiration=%v\n", key, value, expiration)
	}
	return nil
}

// 获取redis key的值
func (c *RedisClient) Get(key string) (string, error) {
	result, err := c.Client.Get(c.Context, key).Result()
	if err != nil {
		if err == redis.Nil {
			if c.IsDebug {
				fmt.Printf("[DEBUG] Redis Get key=%s 不存在\n", key)
			}
			return "", fmt.Errorf("key %s 不存在", key)
		}
		if c.IsDebug {
			fmt.Printf("[ERROR] Redis Get失败 key=%s: %v\n", key, err)
		}
		return "", err
	}
	if c.IsDebug {
		fmt.Printf("[DEBUG] Redis Get成功 key=%s, value=%s\n", key, result)
	}
	return result, nil
}

// 删除redis key
func (c *RedisClient) Del(key string) error {
	err := c.Client.Del(c.Context, key).Err()
	if err != nil {
		if c.IsDebug {
			fmt.Printf("[ERROR] Redis Del失败 key=%s: %v\n", key, err)
		}
		return err
	}
	if c.IsDebug {
		fmt.Printf("[DEBUG] Redis Del成功 key=%s\n", key)
	}
	return nil
}

// 关闭redis连接
func (c *RedisClient) Close() error {
	err := c.Client.Close()
	if err != nil {
		if c.IsDebug {
			fmt.Printf("[ERROR] Redis Close失败: %v\n", err)
		}
		return err
	}
	if c.IsDebug {
		fmt.Printf("[DEBUG] Redis 连接已关闭\n")
	}
	return nil
}

// 设置redis key的过期时间
func (c *RedisClient) Expire(key string, expiration time.Duration) error {
	err := c.Client.Expire(c.Context, key, expiration).Err()
	if err != nil {
		if c.IsDebug {
			fmt.Printf("[ERROR] Redis Expire失败 key=%s: %v\n", key, err)
		}
		return err
	}
	if c.IsDebug {
		fmt.Printf("[DEBUG] Redis Expire成功 key=%s, expiration=%v\n", key, expiration)
	}
	return nil
}

//redis 队列操作

// 向redis队列中添加一个元素
func (c *RedisClient) LPush(key string, values ...interface{}) error {
	err := c.Client.LPush(c.Context, key, values...).Err()
	if err != nil {
		if c.IsDebug {
			fmt.Printf("[ERROR] Redis LPush失败 key=%s: %v\n", key, err)
		}
	}
	return err
}

// 从redis队列中获取一个元素
func (c *RedisClient) RPop(key string) (string, error) {
	result, err := c.Client.RPop(c.Context, key).Result()
	if err != nil && err != redis.Nil {
		if c.IsDebug {
			fmt.Printf("[ERROR] Redis RPop失败 key=%s: %v\n", key, err)
		}
	}
	return result, err
}

//redis 列表操作

// 获取redis列表中指定范围的元素
func (c *RedisClient) LRange(key string, start, stop int64) ([]string, error) {
	return c.Client.LRange(c.Context, key, start, stop).Result()
}

// 获取redis列表中元素的数量
func (c *RedisClient) LLen(key string) (int64, error) {
	return c.Client.LLen(c.Context, key).Result()
}

// redis 哈希操作

// 判断redis哈希表中是否存在指定字段
func (c *RedisClient) HExists(key string, field string) (bool, error) {
	return c.Client.HExists(c.Context, key, field).Result()
}

// 设置redis哈希表中指定字段的值
func (c *RedisClient) HSet(key string, field string, value interface{}) error {
	return c.Client.HSet(c.Context, key, field, value).Err()
}

// 获取redis哈希表中指定字段的值
func (c *RedisClient) HGet(key string, field string) (string, error) {
	return c.Client.HGet(c.Context, key, field).Result()
}

// 获取redis哈希表中所有字段的值
func (c *RedisClient) HGetAll(key string) (map[string]string, error) {
	return c.Client.HGetAll(c.Context, key).Result()
}

// 获取redis哈希表中字段的数量
func (c *RedisClient) HLen(key string) (int64, error) {
	return c.Client.HLen(c.Context, key).Result()
}

//redis 集合操作

// 向redis集合中添加一个元素
func (c *RedisClient) SAdd(key string, members ...interface{}) error {
	return c.Client.SAdd(c.Context, key, members...).Err()
}

// 删除redis集合中指定元素
func (c *RedisClient) SRem(key string, members ...interface{}) error {
	return c.Client.SRem(context.Background(), key, members...).Err()
}

// 判断redis集合中是否存在指定元素
func (c *RedisClient) SIsMember(key string, member interface{}) (bool, error) {
	return c.Client.SIsMember(c.Context, key, member).Result()
}

// 获取redis集合中所有元素
func (c *RedisClient) SMembers(key string) ([]string, error) {
	return c.Client.SMembers(c.Context, key).Result()
}

// 获取redis集合中元素的数量
func (c *RedisClient) SLen(key string) (int64, error) {
	return c.Client.SCard(c.Context, key).Result()
}

//redis 有序集合操作

// 向redis有序集合中添加一个元素
func (c *RedisClient) ZAdd(key string, members ...redis.Z) error {
	return c.Client.ZAdd(c.Context, key, members...).Err()
}

// 获取redis有序集合中指定范围的元素
func (c *RedisClient) ZRange(key string, start, stop int64) ([]string, error) {
	return c.Client.ZRange(c.Context, key, start, stop).Result()
}

// 获取redis有序集合中指定范围的元素及其分数
func (c *RedisClient) ZRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	return c.Client.ZRangeWithScores(c.Context, key, start, stop).Result()
}

// 获取redis有序集合中元素的数量
func (c *RedisClient) ZLen(key string) (int64, error) {
	return c.Client.ZCard(c.Context, key).Result()
}

//redis 发布订阅操作

// 订阅redis频道
func (c *RedisClient) Subscribe(channels ...string) *redis.PubSub {
	return c.Client.Subscribe(c.Context, channels...)
}

// 发布消息到redis频道
func (c *RedisClient) Publish(channel string, message interface{}) error {
	return c.Client.Publish(c.Context, channel, message).Err()
}

//redis 事务操作

// 创建redis事务
func (c *RedisClient) Multi() (redis.Pipeliner, error) {
	return c.Client.TxPipeline(), nil
}

// 执行redis事务
func (c *RedisClient) Exec(pipe redis.Pipeliner) ([]interface{}, error) {
	cmds, err := pipe.Exec(c.Context)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(cmds))
	for i, cmd := range cmds {
		if err := cmd.Err(); err != nil {
			return nil, err
		}
		result[i] = cmd.String()
	}
	return result, nil
}

//redis 位图操作

// 设置redis位图中的指定位置的值
func (c *RedisClient) SetBit(key string, offset int64, value int) (int64, error) {
	return c.Client.SetBit(c.Context, key, offset, value).Result()
}

// 获取redis位图中的指定位置的值
func (c *RedisClient) GetBit(key string, offset int64) (int64, error) {
	return c.Client.GetBit(c.Context, key, offset).Result()
}
