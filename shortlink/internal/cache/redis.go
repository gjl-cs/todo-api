package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"os"
	"shortlink/internal/metrics"
	"strconv"
	"time"
)

const (
	shortLinkCachePrefix = "shortlink:"
	pvCachePrefix        = "shortlink:pv:"

	// 正常短链接缓存 30 分钟。
	shortLinkCacheTTL = 30 * time.Minute

	// 不存在的短码只缓存 60 秒。
	notFoundCacheTTL = 60 * time.Second

	// 负缓存标记：表示这个短码不存在。
	notFoundMarker = "__SHORTLINK_NOT_FOUND__"
)

var Ctx = context.Background()

// Store 定义业务层需要的缓存操作。
type Store interface {
	Get(code string) (string, error)
	Set(code string, longURL string) error
	SetNotFound(code string) error
	IncrementPV(code string) (int64, error)
	GetPV(code string) (int64, error)
}

// RedisStore 是 Redis 缓存的具体实现。
type RedisStore struct {
	client *redis.Client
}

// NewRedis 创建 Redis 客户端。
func NewRedis() *redis.Client {
	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		db = 0
	}

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})
}

// NewRedisStore 创建缓存实例。
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{
		client: client,
	}
}

func (r *RedisStore) Get(code string) (string, error) {
	value, err := r.client.Get(Ctx, "shortlink:"+code).Result()

	if err == nil {
		metrics.CacheHitsTotal.Inc()
	} else {
		metrics.CacheMissesTotal.Inc()
	}

	return value, err
}

// Set 设置正常短链接缓存。
func (r *RedisStore) Set(code string, longURL string) error {
	return Set(r.client, code, longURL)
}

// SetNotFound 缓存短码不存在的结果。
func (r *RedisStore) SetNotFound(code string) error {
	return r.client.Set(
		Ctx,
		shortLinkCachePrefix+code,
		notFoundMarker,
		notFoundCacheTTL,
	).Err()
}

// IncrementPV 增加访问量。
func (r *RedisStore) IncrementPV(code string) (int64, error) {
	return IncrementPV(r.client, code)
}

// GetPV 获取访问量。
func (r *RedisStore) GetPV(code string) (int64, error) {
	return GetPV(r.client, code)
}

// Get 从 Redis 获取短链接。
func Get(client *redis.Client, code string) (string, error) {
	return client.Get(
		Ctx,
		shortLinkCachePrefix+code,
	).Result()
}

// Set 将正常短链接写入 Redis。
func Set(
	client *redis.Client,
	code string,
	longURL string,
) error {
	return client.Set(
		Ctx,
		shortLinkCachePrefix+code,
		longURL,
		shortLinkCacheTTL,
	).Err()
}

// IncrementPV 将访问量加一。
func IncrementPV(
	client *redis.Client,
	code string,
) (int64, error) {
	return client.Incr(
		Ctx,
		pvCachePrefix+code,
	).Result()
}

// GetPV 获取访问量。
func GetPV(
	client *redis.Client,
	code string,
) (int64, error) {
	count, err := client.Get(
		Ctx,
		pvCachePrefix+code,
	).Int64()

	if err == redis.Nil {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return count, nil
}

// IsNotFound 判断缓存值是否代表短码不存在。
func IsNotFound(value string) bool {
	return value == notFoundMarker
}
