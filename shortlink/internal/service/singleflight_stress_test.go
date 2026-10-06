package service

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"shortlink/internal/cache"
)

type singleflightStressRepository struct {
	calls int32
}

func (r *singleflightStressRepository) Create(code, longURL string) error {
	return nil
}

func (r *singleflightStressRepository) GetLongURLByCode(code string) (string, error) {
	atomic.AddInt32(&r.calls, 1)

	// 模拟一次较慢的数据库查询。
	// 只要第一个请求进入这里，其他请求就有充分时间加入 singleflight。
	time.Sleep(100 * time.Millisecond)

	return "https://example.com/singleflight", nil
}

func (r *singleflightStressRepository) Calls() int32 {
	return atomic.LoadInt32(&r.calls)
}

type singleflightStressCache struct {
	mu sync.Mutex

	getCalls int32
	setCalls int32

	value string
}

func (c *singleflightStressCache) Get(code string) (string, error) {
	atomic.AddInt32(&c.getCalls, 1)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.value == "" {
		return "", redis.Nil
	}

	return c.value, nil
}

func (c *singleflightStressCache) Set(code, longURL string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	atomic.AddInt32(&c.setCalls, 1)
	c.value = longURL

	return nil
}

func (c *singleflightStressCache) SetNotFound(code string) error {
	return nil
}

func (c *singleflightStressCache) IncrementPV(code string) (int64, error) {
	return 0, nil
}

func (c *singleflightStressCache) GetPV(code string) (int64, error) {
	return 0, nil
}

func (c *singleflightStressCache) GetCalls() int32 {
	return atomic.LoadInt32(&c.getCalls)
}

func (c *singleflightStressCache) SetCalls() int32 {
	return atomic.LoadInt32(&c.setCalls)
}

func TestGetLongURLSingleflightStress(t *testing.T) {
	const concurrency = 20

	repo := &singleflightStressRepository{}
	cacheStore := &singleflightStressCache{}

	service := &ShortLinkService{
		repo:  repo,
		cache: cacheStore,
	}

	start := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(concurrency)

	results := make([]string, concurrency)
	errs := make([]error, concurrency)

	for i := 0; i < concurrency; i++ {
		index := i

		go func() {
			defer wg.Done()

			<-start

			results[index], errs[index] = service.GetLongURL("test123")
		}()
	}

	close(start)

	wg.Wait()

	for i := 0; i < concurrency; i++ {
		if errs[i] != nil {
			t.Fatalf(
				"goroutine %d returned error: %v",
				i,
				errs[i],
			)
		}

		if results[i] != "https://example.com/singleflight" {
			t.Fatalf(
				"goroutine %d returned unexpected URL: %q",
				i,
				results[i],
			)
		}
	}

	repoCalls := repo.Calls()

	if repoCalls != 1 {
		t.Fatalf(
			"singleflight failed: expected 1 repository call, got %d",
			repoCalls,
		)
	}

	setCalls := cacheStore.SetCalls()

	if setCalls != 1 {
		t.Fatalf(
			"expected 1 cache set, got %d",
			setCalls,
		)
	}

	t.Logf(
		"concurrency=%d repo_calls=%d cache_gets=%d cache_sets=%d",
		concurrency,
		repoCalls,
		cacheStore.GetCalls(),
		setCalls,
	)
}

// 编译期确认 mock 实现了实际 Cache Store 接口。
var _ cache.Store = (*singleflightStressCache)(nil)