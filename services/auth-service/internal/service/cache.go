package services

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"velure-auth-service/internal/metrics"
	"velure-auth-service/internal/model"
)

// CacheEntry representa uma entrada no cache
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// DistributedCache é um cache em memória thread-safe com TTL
type DistributedCache struct {
	data    sync.Map
	cleanup *time.Ticker
	mu      sync.RWMutex
	stop    chan struct{}
}

// NewDistributedCache cria um novo cache distribuído
func NewDistributedCache(cleanupInterval time.Duration) *DistributedCache {
	dc := &DistributedCache{
		cleanup: time.NewTicker(cleanupInterval),
		stop:    make(chan struct{}),
	}

	// Goroutine para limpar entradas expiradas
	go dc.cleanupExpired()

	return dc
}

func (dc *DistributedCache) cleanupExpired() {
	for {
		select {
		case <-dc.cleanup.C:
			now := time.Now()
			dc.data.Range(func(key, value interface{}) bool {
				if entry, ok := value.(*CacheEntry); ok {
					if now.After(entry.ExpiresAt) {
						dc.data.Delete(key)
					}
				}
				return true
			})
		case <-dc.stop:
			return
		}
	}
}

// Set armazena um valor no cache com TTL
func (dc *DistributedCache) Set(key string, value interface{}, ttl time.Duration) {
	entry := &CacheEntry{
		Data:      value,
		ExpiresAt: time.Now().Add(ttl),
	}
	dc.data.Store(key, entry)
}

// Get recupera um valor do cache
func (dc *DistributedCache) Get(key string) (interface{}, bool) {
	value, ok := dc.data.Load(key)
	if !ok {
		metrics.CacheMisses.Inc()
		return nil, false
	}

	entry, ok := value.(*CacheEntry)
	if !ok || time.Now().After(entry.ExpiresAt) {
		dc.data.Delete(key)
		metrics.CacheMisses.Inc()
		return nil, false
	}

	metrics.CacheHits.Inc()
	return entry.Data, true
}

// Delete remove um valor do cache
func (dc *DistributedCache) Delete(key string) {
	dc.data.Delete(key)
}

// GetOrCompute recupera do cache ou computa o valor em paralelo
func (dc *DistributedCache) GetOrCompute(
	ctx context.Context,
	key string,
	ttl time.Duration,
	compute func() (interface{}, error),
) (interface{}, error) {
	// Tentar obter do cache primeiro
	if value, ok := dc.Get(key); ok {
		return value, nil
	}

	// Canal para resultado da computação
	type result struct {
		data interface{}
		err  error
	}
	resultChan := make(chan result, 1)

	// Computar valor em goroutine
	go func() {
		data, err := compute()
		resultChan <- result{data, err}
	}()

	// Aguardar resultado com timeout
	select {
	case res := <-resultChan:
		if res.err == nil {
			dc.Set(key, res.data, ttl)
		}
		return res.data, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// BatchGet recupera múltiplos valores do cache em paralelo
func (dc *DistributedCache) BatchGet(keys []string) map[string]interface{} {
	results := make(map[string]interface{}, len(keys))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			if value, ok := dc.Get(k); ok {
				mu.Lock()
				results[k] = value
				mu.Unlock()
			}
		}(key)
	}

	wg.Wait()
	return results
}

// BatchSet armazena múltiplos valores no cache em paralelo
func (dc *DistributedCache) BatchSet(items map[string]interface{}, ttl time.Duration) {
	var wg sync.WaitGroup

	for key, value := range items {
		wg.Add(1)
		go func(k string, v interface{}) {
			defer wg.Done()
			dc.Set(k, v, ttl)
		}(key, value)
	}

	wg.Wait()
}

// Stop para a goroutine de limpeza
func (dc *DistributedCache) Stop() {
	dc.cleanup.Stop()
	select {
	case <-dc.stop:
	default:
		close(dc.stop)
	}
}

// UserCache é um cache especializado para usuários
type UserCache struct {
	cache *DistributedCache
	ttl   time.Duration
}

func NewUserCache(ttl time.Duration) *UserCache {
	return &UserCache{
		cache: NewDistributedCache(1 * time.Minute),
		ttl:   ttl,
	}
}

func (uc *UserCache) GetUser(userID uint) (*models.User, bool) {
	key := userCacheKey(userID)
	value, ok := uc.cache.Get(key)
	if !ok {
		return nil, false
	}

	user, ok := value.(*models.User)
	return user, ok
}

func (uc *UserCache) SetUser(user *models.User) {
	key := userCacheKey(user.ID)
	uc.cache.Set(key, user, uc.ttl)
}

func (uc *UserCache) GetUserByEmail(email string) (*models.User, bool) {
	key := userEmailCacheKey(email)
	value, ok := uc.cache.Get(key)
	if !ok {
		return nil, false
	}

	user, ok := value.(*models.User)
	return user, ok
}

func (uc *UserCache) SetUserByEmail(user *models.User) {
	key := userEmailCacheKey(user.Email)
	uc.cache.Set(key, user, uc.ttl)
}

func (uc *UserCache) InvalidateUser(userID uint) {
	key := userCacheKey(userID)
	uc.cache.Delete(key)
}

func (uc *UserCache) InvalidateUserByEmail(email string) {
	key := userEmailCacheKey(email)
	uc.cache.Delete(key)
}

func userCacheKey(userID uint) string {
	return "user:" + string(rune(userID))
}

func userEmailCacheKey(email string) string {
	return "user:email:" + email
}

// SessionCache é um cache especializado para sessões
type SessionCache struct {
	cache *DistributedCache
	ttl   time.Duration
}

func NewSessionCache(ttl time.Duration) *SessionCache {
	return &SessionCache{
		cache: NewDistributedCache(1 * time.Minute),
		ttl:   ttl,
	}
}

func (sc *SessionCache) GetSession(token string) (*models.Session, bool) {
	key := sessionCacheKey(token)
	value, ok := sc.cache.Get(key)
	if !ok {
		return nil, false
	}

	session, ok := value.(*models.Session)
	return session, ok
}

func (sc *SessionCache) SetSession(session *models.Session) {
	key := sessionCacheKey(session.AccessToken)
	sc.cache.Set(key, session, sc.ttl)
}

func (sc *SessionCache) InvalidateSession(token string) {
	key := sessionCacheKey(token)
	sc.cache.Delete(key)
}

func sessionCacheKey(token string) string {
	return "session:" + token
}

// SerializeToJSON serializa qualquer struct para JSON em goroutine
func SerializeToJSON(v interface{}) ([]byte, error) {
	type result struct {
		data []byte
		err  error
	}
	resultChan := make(chan result, 1)

	go func() {
		data, err := json.Marshal(v)
		resultChan <- result{data, err}
	}()

	res := <-resultChan
	return res.data, res.err
}

// DeserializeFromJSON desserializa JSON para struct em goroutine
func DeserializeFromJSON(data []byte, v interface{}) error {
	errChan := make(chan error, 1)

	go func() {
		errChan <- json.Unmarshal(data, v)
	}()

	return <-errChan
}
