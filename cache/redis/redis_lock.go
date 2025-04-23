package redis

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// Locker defines the interface for a distributed lock
type Locker interface {
	// Lock attempts to acquire the lock
	Lock(ctx context.Context) error
	// LockWithRetry attempts to acquire the lock with exponential backoff
	LockWithRetry(ctx context.Context, maxRetries int, initialDelay time.Duration) error
	// Unlock releases the lock if it is still held
	Unlock(ctx context.Context) error
	// Refresh extends the lock duration
	Refresh(ctx context.Context) error
	// Close releases any resources held by the lock
	Close(ctx context.Context) error
}

// Lua scripts for atomic operations
const (
	// 解锁脚本：检查并删除锁
	unlockScript = `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end`

	// 刷新脚本：检查并更新过期时间
	refreshScript = `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("pexpire", KEYS[1], ARGV[2])
		else
			return 0
		end`

	// 健康检查脚本：检查锁的状态和剩余时间
	healthCheckScript = `
		local value = redis.call("get", KEYS[1])
		if value == false then
			return {0, 0, 0}  -- key doesn't exist
		end
		local ttl = redis.call("pttl", KEYS[1])
		if value == ARGV[1] then
			return {1, ttl, 1}  -- owned by us
		end
		return {1, ttl, 0}  -- owned by others
	`
)

// LockState represents the current state of the lock
type LockState int32

const (
	// LockStateUnlocked indicates the lock is not held
	LockStateUnlocked LockState = iota
	// LockStateLocked indicates the lock is held
	LockStateLocked
	// LockStateExpired indicates the lock has expired
	LockStateExpired
)

// Common errors that can be returned by the RedisLock
var (
	ErrLockNotObtained   = errors.New("lock not obtained")
	ErrLockNotHeld       = errors.New("lock not held")
	ErrInvalidContext    = errors.New("invalid context")
	ErrLockExpired       = errors.New("lock has expired")
	ErrAlreadyLocked     = errors.New("lock is already held")
	ErrLockKeyRequired   = errors.New("lock key is required")
	ErrLockValueRequired = errors.New("lock value is required")
)

// RedisLock represents a distributed lock using Redis
type RedisLock struct {
	// client is the Redis client instance
	client *redis.Client
	// key is the Redis key used for the lock
	key string
	// value is the unique identifier for this lock instance
	value string
	// expiration is the lock's time-to-live duration
	expiration time.Duration
	state      int32 // atomic
}

// LockHealth represents the health status of a lock
type LockHealth struct {
	Exists    bool          // 锁是否存在
	TTL       time.Duration // 剩余时间
	IsOwner   bool          // 是否是当前持有者
	LastCheck time.Time     // 最后检查时间
}

// NewRedisLock creates a new Redis-based distributed lock
func NewRedisLock(client *redis.Client, key, value string, expiration time.Duration) (*RedisLock, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}
	if key == "" {
		return nil, ErrLockKeyRequired
	}
	if value == "" {
		return nil, ErrLockValueRequired
	}
	if expiration <= 0 {
		return nil, errors.New("expiration must be positive")
	}

	return &RedisLock{
		client:     client,
		key:        key,
		value:      value,
		expiration: expiration,
	}, nil
}

// Lock attempts to acquire the lock
func (l *RedisLock) Lock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Check if already locked
	if !atomic.CompareAndSwapInt32(&l.state, int32(LockStateUnlocked), int32(LockStateLocked)) {
		return ErrAlreadyLocked
	}

	ok, err := l.client.SetNX(ctx, l.key, l.value, l.expiration).Result()

	if err != nil {
		atomic.StoreInt32(&l.state, int32(LockStateUnlocked))
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !ok {
		atomic.StoreInt32(&l.state, int32(LockStateUnlocked))
		return ErrLockNotObtained
	}

	return nil
}

// Unlock releases the lock if it is still held by the current instance
func (l *RedisLock) Unlock(ctx context.Context) error {
	if ctx == nil {
		return ErrInvalidContext
	}

	if !atomic.CompareAndSwapInt32(&l.state, int32(LockStateLocked), int32(LockStateUnlocked)) {
		return ErrLockNotHeld
	}

	result, err := l.client.Eval(ctx, unlockScript, []string{l.key}, l.value).Result()

	if err != nil {
		atomic.StoreInt32(&l.state, int32(LockStateLocked))
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if result.(int64) == 0 {
		atomic.StoreInt32(&l.state, int32(LockStateExpired))
		return ErrLockNotHeld
	}

	return nil
}

// Refresh extends the lock duration if it is still held by the current instance
func (l *RedisLock) Refresh(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	result, err := l.client.Eval(
		ctx,
		refreshScript,
		[]string{l.key},
		l.value,
		l.expiration.Milliseconds(),
	).Result()
	if err != nil {
		return err
	}
	if result.(int64) == 0 {
		return ErrLockNotHeld
	}
	return nil
}

// IsLocked checks if the lock is currently held by any instance
func (l *RedisLock) IsLocked(ctx context.Context) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	exists, err := l.client.Exists(ctx, l.key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

// IsHeldByMe checks if the lock is currently held by this instance
func (l *RedisLock) IsHeldByMe(ctx context.Context) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	value, err := l.client.Get(ctx, l.key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return value == l.value, nil
}

// LockWithRetry attempts to acquire the lock with exponential backoff
func (l *RedisLock) LockWithRetry(ctx context.Context, maxRetries int, initialDelay time.Duration) error {
	if ctx == nil {
		return ErrInvalidContext
	}

	backoff := &ExponentialBackoff{
		InitialDelay: initialDelay,
		MaxDelay:     initialDelay * 10,
		Factor:       2,
	}

	for i := 0; i <= maxRetries; i++ {
		err := l.Lock(ctx)
		if err == nil {
			return nil
		}
		if err != ErrLockNotObtained {
			return err
		}
		if i == maxRetries {
			return ErrLockNotObtained
		}

		delay := backoff.NextBackoff()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue with next retry
		}
	}
	return ErrLockNotObtained
}

// AutoRefresh starts a goroutine to automatically refresh the lock
// It returns a channel that will receive refresh errors and a stop function
func (l *RedisLock) AutoRefresh(ctx context.Context, refreshInterval time.Duration) (chan error, func()) {
	if ctx == nil {
		ctx = context.Background()
	}
	if refreshInterval <= 0 {
		refreshInterval = l.expiration / 3
	}

	resultCh := make(chan error, 1)
	stopCh := make(chan struct{})

	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()
		defer close(resultCh)

		for {
			select {
			case <-ctx.Done():
				resultCh <- ctx.Err()
				return
			case <-stopCh:
				return
			case <-ticker.C:
				if err := l.Refresh(ctx); err != nil {
					resultCh <- err
					return
				}
			}
		}
	}()

	return resultCh, func() { close(stopCh) }
}

// Close implements io.Closer and releases any resources
func (l *RedisLock) Close(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	// Try to unlock if we still hold the lock
	if held, _ := l.IsHeldByMe(ctx); held {
		return l.Unlock(ctx)
	}
	return nil
}

// WithTimeout wraps the lock operation with a timeout
func (l *RedisLock) WithTimeout(timeout time.Duration) *RedisLock {
	l.expiration = timeout
	return l
}

// ForceUnlock unconditionally deletes the lock key
// Use with caution! This method should only be used in emergency situations
func (l *RedisLock) ForceUnlock(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return l.client.Del(ctx, l.key).Err()
}

// String implements fmt.Stringer
func (l *RedisLock) String() string {
	return fmt.Sprintf("RedisLock{key: %s, expiration: %v}", l.key, l.expiration)
}

// Options represents configuration options for RedisLock
type Options struct {
	// RetryCount is the number of retries for lock acquisition
	RetryCount int
	// RetryDelay is the delay between retries
	RetryDelay time.Duration
	// RefreshInterval is the interval for auto-refresh
	RefreshInterval time.Duration
}

// DefaultOptions returns the default options for RedisLock
func DefaultOptions() Options {
	return Options{
		RetryCount:      3,
		RetryDelay:      time.Second,
		RefreshInterval: time.Second * 10,
	}
}

// WithOptions creates a new RedisLock with the given options
func (l *RedisLock) WithOptions(opts Options) *RedisLock {
	if opts.RetryCount <= 0 {
		opts.RetryCount = DefaultOptions().RetryCount
	}
	if opts.RetryDelay <= 0 {
		opts.RetryDelay = DefaultOptions().RetryDelay
	}
	if opts.RefreshInterval <= 0 {
		opts.RefreshInterval = DefaultOptions().RefreshInterval
	}
	return l
}

// ExponentialBackoff implements exponential backoff algorithm
type ExponentialBackoff struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Factor       float64
	current      time.Duration
}

// NextBackoff returns the next backoff delay
func (b *ExponentialBackoff) NextBackoff() time.Duration {
	if b.current == 0 {
		b.current = b.InitialDelay
		return b.current
	}

	b.current = time.Duration(float64(b.current) * b.Factor)
	if b.current > b.MaxDelay {
		b.current = b.MaxDelay
	}
	return b.current
}

// LockWithContext attempts to acquire the lock with a context deadline
func (l *RedisLock) LockWithContext(ctx context.Context) error {
	if ctx == nil {
		return ErrInvalidContext
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		return l.Lock(ctx)
	}

	timeout := time.Until(deadline)
	retryDelay := timeout / 10
	if retryDelay > time.Second {
		retryDelay = time.Second
	}

	return l.LockWithRetry(ctx, int(timeout/retryDelay), retryDelay)
}

// GetTTL returns the remaining time-to-live of the lock
func (l *RedisLock) GetTTL(ctx context.Context) (time.Duration, error) {
	if ctx == nil {
		return 0, ErrInvalidContext
	}

	ttl, err := l.client.TTL(ctx, l.key).Result()
	if err != nil {
		return 0, err
	}
	if ttl < 0 {
		return 0, ErrLockExpired
	}
	return ttl, nil
}

// HealthCheck performs a health check on the lock
func (l *RedisLock) HealthCheck(ctx context.Context) (*LockHealth, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	result, err := l.client.Eval(ctx, healthCheckScript, []string{l.key}, l.value).Result()
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	arr, ok := result.([]interface{})
	if !ok || len(arr) != 3 {
		return nil, errors.New("invalid health check response")
	}

	exists := arr[0].(int64) == 1
	ttl := time.Duration(arr[1].(int64)) * time.Millisecond
	isOwner := arr[2].(int64) == 1

	return &LockHealth{
		Exists:    exists,
		TTL:       ttl,
		IsOwner:   isOwner,
		LastCheck: time.Now(),
	}, nil
}

// Cleanup attempts to clean up any resources associated with the lock
func (l *RedisLock) Cleanup(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// 只有当我们是锁的持有者时才清理
	health, err := l.HealthCheck(ctx)
	if err != nil {
		return fmt.Errorf("cleanup health check failed: %w", err)
	}

	if !health.Exists {
		return nil // 锁已经不存在
	}

	if !health.IsOwner {
		return ErrLockNotHeld
	}

	// 执行清理
	err = l.ForceUnlock(ctx)
	if err != nil {
		return fmt.Errorf("cleanup force unlock failed: %w", err)
	}

	// 重置状态
	atomic.StoreInt32(&l.state, int32(LockStateUnlocked))
	return nil
}

// LockContext is a convenience wrapper that combines Lock with context
type LockContext struct {
	Lock    *RedisLock
	Context context.Context
}

// NewLockContext creates a new LockContext with debug info
func NewLockContext(ctx context.Context, lock *RedisLock) *LockContext {
	return &LockContext{
		Lock:    lock,
		Context: ctx,
	}
}

// Do executes f while holding the lock and automatically releases it after
func (lc *LockContext) Do(f func() error) error {
	if err := lc.Lock.Lock(lc.Context); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer func() {
		unlockErr := lc.Lock.Unlock(lc.Context)
		if unlockErr != nil {
			// Log unlock error but don't return it
			fmt.Printf("failed to release lock: %v\n", unlockErr)
		}
	}()

	return f()
}
