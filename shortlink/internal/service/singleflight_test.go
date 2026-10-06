package service

import (
	"database/sql"
	"errors"
	"github.com/redis/go-redis/v9"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"shortlink/internal/cache"
)

// singleflightTestRepository 用于测试 singleflight 的数据库模拟对象。
type singleflightTestRepository struct {
	calls   int32
	started chan struct{}
	release chan struct{}
	once    sync.Once
	longURL string
}

func (m *singleflightTestRepository) Create(
	code string,
	longURL string,
) error {
	return nil
}

func (m *singleflightTestRepository) GetLongURLByCode(
	code string,
) (string, error) {
	atomic.AddInt32(&m.calls, 1)

	// 通知测试：数据库查询已经开始。
	m.once.Do(func() {
		close(m.started)
	})

	// 暂停数据库查询，让并发请求有机会加入 singleflight。
	<-m.release

	return m.longURL, nil
}

// singleflightTestCache 用于模拟 Redis 缓存未命中。
type singleflightTestCache struct {
	getCalls     int32
	setCalls     int32
	expectedGets int32
	allGetsReady chan struct{}
	once         sync.Once
}

var _ cache.Store = (*singleflightTestCache)(nil)

func (m *singleflightTestCache) Get(code string) (string, error) {
	count := atomic.AddInt32(&m.getCalls, 1)

	// 等所有测试请求都执行到缓存查询阶段。
	if count == m.expectedGets {
		m.once.Do(func() {
			close(m.allGetsReady)
		})
	}

	<-m.allGetsReady

	// 模拟 Redis 缓存未命中。
	return "", redis.Nil
}

func (m *singleflightTestCache) Set(
	code string,
	longURL string,
) error {
	atomic.AddInt32(&m.setCalls, 1)
	return nil
}

func (m *singleflightTestCache) IncrementPV(code string) (int64, error) {
	return 1, nil
}

func (m *singleflightTestCache) GetPV(code string) (int64, error) {
	return 0, nil
}

func TestGetLongURLSingleflight(t *testing.T) {
	const requestCount = 20

	const (
		testCode = "abc123"
		testURL  = "https://example.com/article"
	)

	repo := &singleflightTestRepository{
		started: make(chan struct{}),
		release: make(chan struct{}),
		longURL: testURL,
	}

	cacheStore := &singleflightTestCache{
		expectedGets: requestCount,
		allGetsReady: make(chan struct{}),
	}

	svc := NewShortLinkService(repo, cacheStore)

	results := make(chan error, requestCount)

	// 同时发起 20 个请求，查询相同的短码。
	for i := 0; i < requestCount; i++ {
		go func() {
			gotURL, err := svc.GetLongURL(testCode)
			if err != nil {
				results <- err
				return
			}

			if gotURL != testURL {
				results <- errors.New("unexpected long URL: " + gotURL)
				return
			}

			results <- nil
		}()
	}

	// 等待所有请求都执行过缓存查询，并确认数据库回源已开始。
	select {
	case <-repo.started:
	case <-time.After(3 * time.Second):
		close(repo.release)
		t.Fatal("timeout waiting for database query")
	}

	// 给其他请求时间加入同一个 singleflight 操作。
	time.Sleep(100 * time.Millisecond)

	// 放行数据库查询。
	close(repo.release)

	// 检查所有请求是否都成功。
	for i := 0; i < requestCount; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Errorf("request failed: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("timeout waiting for concurrent requests")
		}
	}

	// 核心断言：同一个短码的并发回源只执行一次。
	if got := atomic.LoadInt32(&repo.calls); got != 1 {
		t.Errorf("expected 1 database query, got %d", got)
	}

	// 每个请求先查询一次缓存，singleflight 回调还会额外查询一次。
	expectedCacheGets := int32(requestCount + 1)

	if got := atomic.LoadInt32(&cacheStore.getCalls); got != expectedCacheGets {
		t.Errorf(
			"expected %d cache gets, got %d",
			expectedCacheGets,
			got,
		)
	}

	// 20 个请求各自查询一次缓存。
	// singleflight 回调还会额外检查一次缓存。
	expectedCacheGets = int32(requestCount + 1)

	if got := atomic.LoadInt32(&cacheStore.getCalls); got != expectedCacheGets {
		t.Errorf(
			"expected %d cache gets, got %d",
			expectedCacheGets,
			got,
		)
	}
}

// SetNotFound 模拟缓存不存在的短链接。
func (c *singleflightTestCache) SetNotFound(code string) error {
	return nil
}

// negativeSingleflightRepository 模拟不存在短码的数据库查询。
type negativeSingleflightRepository struct {
	calls   int32
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (m *negativeSingleflightRepository) Create(
	code string,
	longURL string,
) error {
	return nil
}

func (m *negativeSingleflightRepository) GetLongURLByCode(
	code string,
) (string, error) {
	atomic.AddInt32(&m.calls, 1)

	m.once.Do(func() {
		close(m.started)
	})

	// 暂停数据库查询，让其他并发请求有机会加入 singleflight。
	<-m.release

	return "", sql.ErrNoRows
}

// negativeSingleflightCache 模拟 Redis 缓存未命中及负缓存写入。
type negativeSingleflightCache struct {
	getCalls         int32
	setNotFoundCalls int32
	expectedGets     int32
	allGetsReady     chan struct{}
	once             sync.Once
}

var _ cache.Store = (*negativeSingleflightCache)(nil)

func (m *negativeSingleflightCache) Get(
	code string,
) (string, error) {
	count := atomic.AddInt32(&m.getCalls, 1)

	if count == m.expectedGets {
		m.once.Do(func() {
			close(m.allGetsReady)
		})
	}

	// 等所有并发请求完成第一次缓存查询。
	<-m.allGetsReady

	return "", redis.Nil
}

func (m *negativeSingleflightCache) Set(
	code string,
	longURL string,
) error {
	return nil
}

func (m *negativeSingleflightCache) SetNotFound(
	code string,
) error {
	atomic.AddInt32(&m.setNotFoundCalls, 1)
	return nil
}

func (m *negativeSingleflightCache) IncrementPV(
	code string,
) (int64, error) {
	return 1, nil
}

func (m *negativeSingleflightCache) GetPV(
	code string,
) (int64, error) {
	return 0, nil
}

// TestGetLongURLNotFoundSingleflight 验证并发访问不存在的短码时，
// 数据库查询和负缓存写入都只发生一次。
func TestGetLongURLNotFoundSingleflight(t *testing.T) {
	const requestCount = 20

	repo := &negativeSingleflightRepository{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}

	cacheMock := &negativeSingleflightCache{
		expectedGets: int32(requestCount),
		allGetsReady: make(chan struct{}),
	}

	svc := NewShortLinkService(repo, cacheMock)

	errs := make([]error, requestCount)

	var wg sync.WaitGroup
	wg.Add(requestCount)

	for i := 0; i < requestCount; i++ {
		go func(index int) {
			defer wg.Done()

			_, err := svc.GetLongURL("missing-concurrent")
			errs[index] = err
		}(i)
	}

	// 等待所有请求完成第一次缓存查询。
	select {
	case <-cacheMock.allGetsReady:
	case <-time.After(5 * time.Second):
		close(repo.release)
		t.Fatal("timed out waiting for concurrent cache queries")
	}

	// 等待 singleflight 中的数据库查询开始。
	select {
	case <-repo.started:
	case <-time.After(5 * time.Second):
		close(repo.release)
		t.Fatal("timed out waiting for database query")
	}

	// 给其他请求加入同一轮 singleflight 的机会。
	time.Sleep(100 * time.Millisecond)

	close(repo.release)
	wg.Wait()

	// 每个请求都应该得到短码不存在的错误。
	for i, err := range errs {
		if !errors.Is(err, ErrShortLinkNotFound) {
			t.Errorf(
				"request %d: expected ErrShortLinkNotFound, got %v",
				i,
				err,
			)
		}
	}

	// 同一轮并发请求只允许一次数据库回源。
	if got := atomic.LoadInt32(&repo.calls); got != 1 {
		t.Errorf("expected 1 database query, got %d", got)
	}

	// 数据库确认短码不存在后，只写入一次负缓存。
	if got := atomic.LoadInt32(&cacheMock.setNotFoundCalls); got != 1 {
		t.Errorf("expected 1 negative cache write, got %d", got)
	}
}
