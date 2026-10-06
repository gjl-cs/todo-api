package service

import (
	"database/sql"
	"errors"
	"fmt"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"golang.org/x/sync/singleflight"

	"shortlink/internal/cache"
	"shortlink/internal/utils"
)

// ErrShortLinkNotFound 表示短链接不存在。
var ErrShortLinkNotFound = errors.New("short link not found")

// ShortLinkRepository 定义短链接业务需要的数据库操作。
type ShortLinkRepository interface {
	Create(code string, longURL string) error
	GetLongURLByCode(code string) (string, error)
}

// ShortLinkService 处理短链接业务。
type ShortLinkService struct {
	repo  ShortLinkRepository
	cache cache.Store

	// group 合并同一个短码的并发回源请求。
	group singleflight.Group
}

// NewShortLinkService 创建短链接业务服务。
func NewShortLinkService(
	repo ShortLinkRepository,
	cacheStore cache.Store,
) *ShortLinkService {
	return &ShortLinkService{
		repo:  repo,
		cache: cacheStore,
	}
}

// CreateShortLink 创建短链接。
func (s *ShortLinkService) CreateShortLink(url string) (string, error) {
	// 创建短链接之前先校验原始 URL。
	if err := utils.ValidateLongURL(url); err != nil {
		return "", err
	}

	const maxRetries = 5

	for i := 0; i < maxRetries; i++ {
		code, err := utils.GenerateShortCode(6)
		if err != nil {
			return "", fmt.Errorf("generate short code failed: %w", err)
		}

		err = s.repo.Create(code, url)
		if err == nil {
			return code, nil
		}

		// 短码发生唯一键冲突时，重新生成。
		var mysqlErr *mysqlDriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			continue
		}

		return "", fmt.Errorf("create short link failed: %w", err)
	}

	return "", fmt.Errorf(
		"failed to generate unique short code after %d attempts",
		maxRetries,
	)
}

// GetLongURL 根据短码获取原始 URL。
func (s *ShortLinkService) GetLongURL(code string) (string, error) {
	// 第一步：先查 Redis。
	longURL, err := s.cache.Get(code)

	// 命中缓存。
	if err == nil {
		// 如果命中的是负缓存，说明短码不存在。
		if cache.IsNotFound(longURL) {
			return "", ErrShortLinkNotFound
		}

		// 正常缓存直接返回。
		return longURL, nil
	}

	// Redis 查询失败时，继续尝试查询 MySQL。
	// 例如缓存中没有这个 Key，或者 Redis 暂时不可用。

	// 第二步：使用 singleflight 合并相同短码的并发查询。
	result, err, _ := s.group.Do(code, func() (interface{}, error) {
		// 再检查一次缓存。
		// 其他并发请求可能已经完成了数据库查询并写入缓存。
		cachedURL, cacheErr := s.cache.Get(code)

		if cacheErr == nil {
			if cache.IsNotFound(cachedURL) {
				return "", ErrShortLinkNotFound
			}

			return cachedURL, nil
		}

		// 第三步：缓存没有结果，查询 MySQL。
		longURL, repoErr := s.repo.GetLongURLByCode(code)

		if repoErr != nil {
			// 只有明确确认记录不存在，才写入负缓存。
			if errors.Is(repoErr, sql.ErrNoRows) {
				if cacheErr := s.cache.SetNotFound(code); cacheErr != nil {
					fmt.Println("redis negative cache set failed:", cacheErr)
				}

				return "", ErrShortLinkNotFound
			}

			// 数据库故障等其他错误，不能缓存成“不存在”。
			return "", fmt.Errorf("get short link failed: %w", repoErr)
		}

		// 第四步：数据库查询成功，写入正常缓存。
		if cacheErr := s.cache.Set(code, longURL); cacheErr != nil {
			fmt.Println("redis cache set failed:", cacheErr)
		}

		return longURL, nil
	})

	if err != nil {
		return "", err
	}

	longURL, ok := result.(string)
	if !ok {
		return "", errors.New("unexpected singleflight result type")
	}

	return longURL, nil
}

// RecordVisit 记录一次短链接访问。
func (s *ShortLinkService) RecordVisit(code string) error {
	_, err := s.cache.IncrementPV(code)
	return err
}

// GetPV 查询短链接的访问次数。
func (s *ShortLinkService) GetPV(code string) (int64, error) {
	return s.cache.GetPV(code)
}
