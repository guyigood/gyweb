package redisutils

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient 是 Redis 客户端
type RedisClient struct {
	Client  *redis.Client
	Context context.Context
}

func NewRedisClient(addr string, password string, db int) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	ctx := context.Background()

	return &RedisClient{Client: client, Context: ctx}
}

// 判断redis key是否存在
func (c *RedisClient) Exists(key string) (bool, error) {
	n, err := c.Client.Exists(context.Background(), key).Result()
	return n > 0, err
}

// 设置redis key的值
func (c *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return c.Client.Set(context.Background(), key, value, expiration).Err()
}

// 获取redis key的值
func (c *RedisClient) Get(key string) (string, error) {
	return c.Client.Get(context.Background(), key).Result()
}

// 删除redis key
func (c *RedisClient) Del(key string) error {
	return c.Client.Del(context.Background(), key).Err()
}

// 关闭redis连接
func (c *RedisClient) Close() error {
	return c.Client.Close()
}

//redis 队列操作

// 向redis队列中添加一个元素
func (c *RedisClient) LPush(key string, values ...interface{}) error {
	return c.Client.LPush(context.Background(), key, values...).Err()
}

// 从redis队列中获取一个元素
func (c *RedisClient) RPop(key string) (string, error) {
	return c.Client.RPop(context.Background(), key).Result()
}

//redis 列表操作

// 获取redis列表中指定范围的元素
func (c *RedisClient) LRange(key string, start, stop int64) ([]string, error) {
	return c.Client.LRange(context.Background(), key, start, stop).Result()
}

// 获取redis列表中元素的数量
func (c *RedisClient) LLen(key string) (int64, error) {
	return c.Client.LLen(context.Background(), key).Result()
}

// redis 哈希操作

// 判断redis哈希表中是否存在指定字段
func (c *RedisClient) HExists(key string, field string) (bool, error) {
	return c.Client.HExists(context.Background(), key, field).Result()
}

// 设置redis哈希表中指定字段的值
func (c *RedisClient) HSet(key string, field string, value interface{}) error {
	return c.Client.HSet(context.Background(), key, field, value).Err()
}

// 获取redis哈希表中指定字段的值
func (c *RedisClient) HGet(key string, field string) (string, error) {
	return c.Client.HGet(context.Background(), key, field).Result()
}

// 获取redis哈希表中所有字段的值
func (c *RedisClient) HGetAll(key string) (map[string]string, error) {
	return c.Client.HGetAll(context.Background(), key).Result()
}

// 获取redis哈希表中字段的数量
func (c *RedisClient) HLen(key string) (int64, error) {
	return c.Client.HLen(context.Background(), key).Result()
}

//redis 集合操作

// 向redis集合中添加一个元素
func (c *RedisClient) SAdd(key string, members ...interface{}) error {
	return c.Client.SAdd(context.Background(), key, members...).Err()
}

// 删除redis集合中指定元素
func (c *RedisClient) SRem(key string, members ...interface{}) error {
	return c.Client.SRem(context.Background(), key, members...).Err()
}

// 判断redis集合中是否存在指定元素
func (c *RedisClient) SIsMember(key string, member interface{}) (bool, error) {
	return c.Client.SIsMember(context.Background(), key, member).Result()
}

// 获取redis集合中所有元素
func (c *RedisClient) SMembers(key string) ([]string, error) {
	return c.Client.SMembers(context.Background(), key).Result()
}

// 获取redis集合中元素的数量
func (c *RedisClient) SLen(key string) (int64, error) {
	return c.Client.SCard(context.Background(), key).Result()
}

//redis 有序集合操作

// 向redis有序集合中添加一个元素
func (c *RedisClient) ZAdd(key string, members ...redis.Z) error {
	return c.Client.ZAdd(context.Background(), key, members...).Err()
}

// 获取redis有序集合中指定范围的元素
func (c *RedisClient) ZRange(key string, start, stop int64) ([]string, error) {
	return c.Client.ZRange(context.Background(), key, start, stop).Result()
}

// 获取redis有序集合中指定范围的元素及其分数
func (c *RedisClient) ZRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	return c.Client.ZRangeWithScores(context.Background(), key, start, stop).Result()
}

// 获取redis有序集合中元素的数量
func (c *RedisClient) ZLen(key string) (int64, error) {
	return c.Client.ZCard(context.Background(), key).Result()
}

//redis 发布订阅操作

// 订阅redis频道
func (c *RedisClient) Subscribe(channels ...string) *redis.PubSub {
	return c.Client.Subscribe(context.Background(), channels...)
}

// 发布消息到redis频道
func (c *RedisClient) Publish(channel string, message interface{}) error {
	return c.Client.Publish(context.Background(), channel, message).Err()
}

//redis 事务操作

// 创建redis事务
func (c *RedisClient) Multi() (redis.Pipeliner, error) {
	return c.Client.TxPipeline(), nil
}

// 执行redis事务
func (c *RedisClient) Exec(pipe redis.Pipeliner) ([]interface{}, error) {
	cmds, err := pipe.Exec(context.Background())
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
	return c.Client.SetBit(context.Background(), key, offset, value).Result()
}

// 获取redis位图中的指定位置的值
func (c *RedisClient) GetBit(key string, offset int64) (int64, error) {
	return c.Client.GetBit(context.Background(), key, offset).Result()
}
