package service

import (
	"database/sql"
	"errors"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"testing"
)

// ==================== Mock Repository ====================

// mockShortLinkRepository 模拟数据库操作。
type mockShortLinkRepository struct {
	calls    int
	errors   []error
	longURL  string
	getErr   error
	getCalls int
}

func (m *mockShortLinkRepository) Create(code string, longURL string) error {
	m.calls++

	index := m.calls - 1
	if index < len(m.errors) {
		return m.errors[index]
	}

	return nil
}

func (m *mockShortLinkRepository) GetLongURLByCode(code string) (string, error) {
	m.getCalls++

	if m.getErr != nil {
		return "", m.getErr
	}

	return m.longURL, nil
}

// ==================== Mock Cache ====================

// mockCache 模拟 Redis 缓存，不需要启动真实 Redis。
type mockCache struct {
	value        string
	getErr       error
	getCalls     int
	setCalls     int
	setValue     string
	setErr       error
	incrementPV  int64
	incrementErr error
	pv           int64
	getPVErr     error
}

func (m *mockCache) Get(code string) (string, error) {
	m.getCalls++
	return m.value, m.getErr
}

func (m *mockCache) Set(code string, value string) error {
	m.setCalls++
	m.setValue = value
	return m.setErr
}

// SetNotFound 模拟缓存不存在的短链接。
func (m *mockCache) SetNotFound(code string) error {
	m.setCalls++

	if m.setErr != nil {
		return m.setErr
	}

	m.value = "__SHORTLINK_NOT_FOUND__"
	m.getErr = nil

	return nil
}

func (m *mockCache) IncrementPV(code string) (int64, error) {
	if m.incrementErr != nil {
		return 0, m.incrementErr
	}

	m.incrementPV++
	return m.incrementPV, nil
}

func (m *mockCache) GetPV(code string) (int64, error) {
	if m.getPVErr != nil {
		return 0, m.getPVErr
	}

	return m.pv, nil
}

// ==================== CreateShortLink Tests ====================

// 第一次发生短码冲突，第二次创建成功。
func TestCreateShortLinkRetryOnDuplicate(t *testing.T) {
	repo := &mockShortLinkRepository{
		errors: []error{
			&mysqlDriver.MySQLError{
				Number:  1062,
				Message: "Duplicate entry",
			},
			nil,
		},
	}

	svc := NewShortLinkService(repo, nil)

	code, err := svc.CreateShortLink("https://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(code) != 6 {
		t.Fatalf("expected code length 6, got %d", len(code))
	}

	if repo.calls != 2 {
		t.Fatalf("expected 2 create calls, got %d", repo.calls)
	}
}

// 普通数据库错误不应该反复重试。
func TestCreateShortLinkStopOnDatabaseError(t *testing.T) {
	repo := &mockShortLinkRepository{
		errors: []error{
			errors.New("database connection failed"),
		},
	}

	svc := NewShortLinkService(repo, nil)

	_, err := svc.CreateShortLink("https://example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if repo.calls != 1 {
		t.Fatalf("expected 1 create call, got %d", repo.calls)
	}
}

// 连续发生 5 次短码冲突后应该返回错误。
func TestCreateShortLinkMaxRetries(t *testing.T) {
	duplicateErr := &mysqlDriver.MySQLError{
		Number:  1062,
		Message: "Duplicate entry",
	}

	repo := &mockShortLinkRepository{
		errors: []error{
			duplicateErr,
			duplicateErr,
			duplicateErr,
			duplicateErr,
			duplicateErr,
		},
	}

	svc := NewShortLinkService(repo, nil)

	_, err := svc.CreateShortLink("https://example.com")
	if err == nil {
		t.Fatal("expected error after maximum retries")
	}

	if repo.calls != 5 {
		t.Fatalf("expected 5 create calls, got %d", repo.calls)
	}
}

// ==================== GetLongURL Tests ====================

// 缓存命中时，直接返回缓存中的 URL，不查询数据库。
func TestGetLongURLCacheHit(t *testing.T) {
	cacheMock := &mockCache{
		value: "https://example.com",
	}

	repo := &mockShortLinkRepository{
		longURL: "https://database.example.com",
	}

	svc := NewShortLinkService(repo, cacheMock)

	got, err := svc.GetLongURL("abc123")
	if err != nil {
		t.Fatalf("获取原始 URL 失败: %v", err)
	}

	if got != "https://example.com" {
		t.Fatalf("原始 URL 错误: got %q", got)
	}

	if repo.getCalls != 0 {
		t.Fatalf(
			"缓存命中时不应该查询数据库，实际查询 %d 次",
			repo.getCalls,
		)
	}

	if cacheMock.getCalls != 1 {
		t.Fatalf(
			"预期查询缓存 1 次，实际查询 %d 次",
			cacheMock.getCalls,
		)
	}
}

// 缓存未命中时，查询数据库，并尝试将结果写回缓存。
func TestGetLongURLCacheMiss(t *testing.T) {
	cacheMock := &mockCache{
		getErr: redis.Nil,
	}

	repo := &mockShortLinkRepository{
		longURL: "https://example.com",
	}

	svc := NewShortLinkService(repo, cacheMock)

	got, err := svc.GetLongURL("abc123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got != "https://example.com" {
		t.Fatalf("expected database URL, got %q", got)
	}

	if repo.getCalls != 1 {
		t.Fatalf("expected 1 database query, got %d", repo.getCalls)
	}

	if cacheMock.setCalls != 1 {
		t.Fatalf("expected 1 cache write, got %d", cacheMock.setCalls)
	}

	if cacheMock.setValue != "https://example.com" {
		t.Fatalf("expected cached URL to be written back, got %q", cacheMock.setValue)
	}
}

// 数据库查询失败时，应该返回错误。
func TestGetLongURLDatabaseError(t *testing.T) {
	cacheMock := &mockCache{
		getErr: redis.Nil,
	}

	repo := &mockShortLinkRepository{
		getErr: errors.New("database query failed"),
	}

	svc := NewShortLinkService(repo, cacheMock)

	_, err := svc.GetLongURL("abc123")
	if err == nil {
		t.Fatal("expected database error, got nil")
	}

	if cacheMock.setCalls != 0 {
		t.Fatalf("expected no cache write, got %d calls", cacheMock.setCalls)
	}
}

// ==================== PV Tests ====================

// 验证访问次数递增。
func TestRecordVisit(t *testing.T) {
	cacheMock := &mockCache{}

	svc := NewShortLinkService(&mockShortLinkRepository{}, cacheMock)

	err := svc.RecordVisit("abc123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cacheMock.incrementPV != 1 {
		t.Fatalf("expected PV to increment to 1, got %d", cacheMock.incrementPV)
	}
}

// 验证访问次数查询。
func TestGetPV(t *testing.T) {
	cacheMock := &mockCache{
		pv: 10,
	}

	svc := NewShortLinkService(&mockShortLinkRepository{}, cacheMock)

	got, err := svc.GetPV("abc123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got != 10 {
		t.Fatalf("expected PV 10, got %d", got)
	}
}
func TestCreateShortLinkInvalidURL(t *testing.T) {
	repo := &mockShortLinkRepository{}

	svc := NewShortLinkService(repo, nil)

	_, err := svc.CreateShortLink("not-a-valid-url")
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}

	if repo.calls != 0 {
		t.Fatalf(
			"expected no database calls for invalid URL, got %d",
			repo.calls,
		)
	}
}
func TestGetLongURLCacheErrorFallsBackToDatabase(t *testing.T) {
	cacheMock := &mockCache{
		getErr: errors.New("redis connection failed"),
	}

	repo := &mockShortLinkRepository{
		longURL: "https://example.com",
	}

	svc := NewShortLinkService(repo, cacheMock)

	got, err := svc.GetLongURL("abc123")
	if err != nil {
		t.Fatalf("expected database fallback to succeed, got %v", err)
	}

	if got != "https://example.com" {
		t.Fatalf("expected database URL, got %q", got)
	}

	if repo.getCalls != 1 {
		t.Fatalf("expected 1 database query, got %d", repo.getCalls)
	}
}
func TestGetLongURLNotFound(t *testing.T) {
	cacheMock := &mockCache{
		getErr: redis.Nil,
	}

	repo := &mockShortLinkRepository{
		getErr: sql.ErrNoRows,
	}

	svc := NewShortLinkService(repo, cacheMock)

	// 第一次请求：缓存未命中，查询数据库，确认短码不存在。
	_, err := svc.GetLongURL("missing")
	if !errors.Is(err, ErrShortLinkNotFound) {
		t.Fatalf("expected ErrShortLinkNotFound, got %v", err)
	}

	if repo.getCalls != 1 {
		t.Fatalf("expected 1 database query after first request, got %d", repo.getCalls)
	}

	if cacheMock.setCalls != 1 {
		t.Fatalf("expected 1 negative cache write, got %d", cacheMock.setCalls)
	}

	// 第二次请求：应该命中负缓存，不再查询数据库。
	_, err = svc.GetLongURL("missing")
	if !errors.Is(err, ErrShortLinkNotFound) {
		t.Fatalf("expected ErrShortLinkNotFound on second request, got %v", err)
	}

	if repo.getCalls != 1 {
		t.Fatalf("expected database query count to remain 1, got %d", repo.getCalls)
	}

	if cacheMock.setCalls != 1 {
		t.Fatalf("expected negative cache write count to remain 1, got %d", cacheMock.setCalls)
	}
}

func TestGetLongURLDatabaseFailure(t *testing.T) {
	cacheMock := &mockCache{
		getErr: redis.Nil,
	}

	databaseErr := errors.New("database connection failed")

	repo := &mockShortLinkRepository{
		getErr: databaseErr,
	}

	svc := NewShortLinkService(repo, cacheMock)

	_, err := svc.GetLongURL("abc123")

	if err == nil {
		t.Fatal("expected database error, got nil")
	}

	if errors.Is(err, ErrShortLinkNotFound) {
		t.Fatal("database failure should not be treated as not found")
	}

	if !errors.Is(err, databaseErr) {
		t.Fatalf(
			"expected wrapped database error, got %v",
			err,
		)
	}
}
func TestGetLongURLCacheSetFailure(t *testing.T) {
	cacheMock := &mockCache{
		getErr: errors.New("redis read failed"),
		setErr: errors.New("redis write failed"),
	}

	repo := &mockShortLinkRepository{
		longURL: "https://example.com",
	}

	svc := NewShortLinkService(repo, cacheMock)

	got, err := svc.GetLongURL("abc123")
	if err != nil {
		t.Fatalf("缓存写入失败不应该影响查询: %v", err)
	}

	if got != "https://example.com" {
		t.Fatalf("原始 URL 错误: got %q", got)
	}

	if repo.getCalls != 1 {
		t.Fatalf(
			"预期查询数据库 1 次，实际查询 %d 次",
			repo.getCalls,
		)
	}

	if cacheMock.setCalls != 1 {
		t.Fatalf(
			"预期尝试写入缓存 1 次，实际写入 %d 次",
			cacheMock.setCalls,
		)
	}
}
func TestGetLongURLNotFoundCacheSetFailure(t *testing.T) {
	cacheMock := &mockCache{
		getErr: redis.Nil,
		setErr: errors.New("redis write failed"),
	}

	repo := &mockShortLinkRepository{
		getErr: sql.ErrNoRows,
	}

	svc := NewShortLinkService(repo, cacheMock)

	_, err := svc.GetLongURL("missing")

	if !errors.Is(err, ErrShortLinkNotFound) {
		t.Fatalf("expected ErrShortLinkNotFound, got %v", err)
	}

	if repo.getCalls != 1 {
		t.Fatalf("expected 1 database query, got %d", repo.getCalls)
	}

	if cacheMock.setCalls != 1 {
		t.Fatalf("expected 1 negative cache write attempt, got %d", cacheMock.setCalls)
	}
}
