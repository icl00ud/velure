package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"velure-auth-service/internal/model"
)

func TestDistributedCache_SetGetAndExpire(t *testing.T) {
	cache := NewDistributedCache(5 * time.Millisecond)
	defer cache.Stop()

	cache.Set("key", "value", 10*time.Millisecond)

	val, ok := cache.Get("key")
	if !ok || val.(string) != "value" {
		t.Fatalf("expected cached value, got %v (ok=%v)", val, ok)
	}

	time.Sleep(20 * time.Millisecond)

	if _, ok := cache.Get("key"); ok {
		t.Fatalf("expected entry to be expired and cleaned up")
	}
}

func TestDistributedCache_GetOrCompute(t *testing.T) {
	cache := NewDistributedCache(50 * time.Millisecond)
	defer cache.Stop()

	calls := 0
	compute := func() (interface{}, error) {
		calls++
		return "computed", nil
	}

	val, err := cache.GetOrCompute(context.Background(), "key", time.Second, compute)
	if err != nil || val.(string) != "computed" {
		t.Fatalf("unexpected result from GetOrCompute: %v err=%v", val, err)
	}

	// Second call should hit cache and not execute compute again
	val, err = cache.GetOrCompute(context.Background(), "key", time.Second, func() (interface{}, error) {
		return nil, errors.New("should not be called")
	})
	if err != nil || val.(string) != "computed" {
		t.Fatalf("expected cached result, got %v err=%v", val, err)
	}

	if calls != 1 {
		t.Fatalf("compute should have been called once, got %d", calls)
	}
}

func TestDistributedCache_GetOrComputeCanceled(t *testing.T) {
	cache := NewDistributedCache(50 * time.Millisecond)
	defer cache.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := cache.GetOrCompute(ctx, "key", time.Second, func() (interface{}, error) {
		time.Sleep(20 * time.Millisecond)
		return "value", nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation error, got %v", err)
	}
}

func TestDistributedCache_BatchOperations(t *testing.T) {
	cache := NewDistributedCache(50 * time.Millisecond)
	defer cache.Stop()

	items := map[string]interface{}{
		"a": 1,
		"b": "two",
	}

	cache.BatchSet(items, time.Second)

	results := cache.BatchGet([]string{"a", "b", "missing"})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results["a"].(int) != 1 || results["b"].(string) != "two" {
		t.Fatalf("unexpected batch results: %#v", results)
	}
}

func TestDistributedCache_Delete(t *testing.T) {
	cache := NewDistributedCache(50 * time.Millisecond)
	defer cache.Stop()

	cache.Set("key", "value", time.Second)
	cache.Delete("key")

	if _, ok := cache.Get("key"); ok {
		t.Fatalf("expected key to be deleted")
	}
}

func TestDistributedCache_Miss(t *testing.T) {
	cache := NewDistributedCache(50 * time.Millisecond)
	defer cache.Stop()

	if _, ok := cache.Get("missing"); ok {
		t.Fatalf("expected cache miss for unknown key")
	}
}

func TestDistributedCache_InvalidEntry(t *testing.T) {
	cache := NewDistributedCache(50 * time.Millisecond)
	defer cache.Stop()

	cache.data.Store("bad", "value")

	if _, ok := cache.Get("bad"); ok {
		t.Fatalf("expected cache miss for invalid entry type")
	}
}

func TestUserCacheOperations(t *testing.T) {
	userCache := NewUserCache(100 * time.Millisecond)
	defer userCache.cache.Stop()

	user := &models.User{ID: 1, Email: "user@example.com", Name: "User"}

	userCache.SetUser(user)
	cachedUser, ok := userCache.GetUser(user.ID)
	if !ok || cachedUser.Email != user.Email {
		t.Fatalf("expected cached user by ID")
	}

	userCache.SetUserByEmail(user)
	cachedByEmail, ok := userCache.GetUserByEmail(user.Email)
	if !ok || cachedByEmail.ID != user.ID {
		t.Fatalf("expected cached user by email")
	}

	userCache.InvalidateUser(user.ID)
	if _, ok := userCache.GetUser(user.ID); ok {
		t.Fatalf("expected user cache by ID to be invalidated")
	}

	userCache.InvalidateUserByEmail(user.Email)
	if _, ok := userCache.GetUserByEmail(user.Email); ok {
		t.Fatalf("expected user cache by email to be invalidated")
	}
}

func TestSessionCacheOperations(t *testing.T) {
	sessionCache := NewSessionCache(100 * time.Millisecond)
	defer sessionCache.cache.Stop()

	session := &models.Session{
		UserID:      1,
		AccessToken: "access",
	}

	sessionCache.SetSession(session)

	cachedSession, ok := sessionCache.GetSession(session.AccessToken)
	if !ok || cachedSession.UserID != session.UserID {
		t.Fatalf("expected cached session")
	}

	sessionCache.InvalidateSession(session.AccessToken)
	if _, ok := sessionCache.GetSession(session.AccessToken); ok {
		t.Fatalf("expected session to be invalidated")
	}
}

func TestSerializeDeserializeJSON(t *testing.T) {
	payload := map[string]string{"hello": "world"}

	data, err := SerializeToJSON(payload)
	if err != nil {
		t.Fatalf("SerializeToJSON error: %v", err)
	}

	var decoded map[string]string
	if err := DeserializeFromJSON(data, &decoded); err != nil {
		t.Fatalf("DeserializeFromJSON error: %v", err)
	}

	if decoded["hello"] != "world" {
		t.Fatalf("unexpected decoded payload: %v", decoded)
	}

	// Validate it produces valid JSON
	if !json.Valid(data) {
		t.Fatalf("expected valid json from SerializeToJSON")
	}
}
