package models

import (
	"testing"
	"time"
)

func TestUser_ToResponse(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:        1,
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  "hashed_password",
		CreatedAt: now,
		UpdatedAt: now,
	}

	response := user.ToResponse()

	if response.ID != user.ID {
		t.Errorf("Expected ID %d, got %d", user.ID, response.ID)
	}
	if response.Name != user.Name {
		t.Errorf("Expected Name %s, got %s", user.Name, response.Name)
	}
	if response.Email != user.Email {
		t.Errorf("Expected Email %s, got %s", user.Email, response.Email)
	}
	if response.CreatedAt != user.CreatedAt {
		t.Errorf("Expected CreatedAt %v, got %v", user.CreatedAt, response.CreatedAt)
	}
	if response.UpdatedAt != user.UpdatedAt {
		t.Errorf("Expected UpdatedAt %v, got %v", user.UpdatedAt, response.UpdatedAt)
	}
}

func TestUser_BeforeCreate(t *testing.T) {
	user := &User{
		Name:  "Test User",
		Email: "test@example.com",
	}

	// BeforeCreate should set CreatedAt and UpdatedAt
	err := user.BeforeCreate(nil)
	if err != nil {
		t.Errorf("BeforeCreate() returned unexpected error: %v", err)
	}

	if user.CreatedAt.IsZero() {
		t.Error("BeforeCreate() should set CreatedAt, but it's zero")
	}
	if user.UpdatedAt.IsZero() {
		t.Error("BeforeCreate() should set UpdatedAt, but it's zero")
	}

	// Verify CreatedAt and UpdatedAt are recent (within 1 second)
	now := time.Now()
	if now.Sub(user.CreatedAt) > time.Second {
		t.Error("CreatedAt should be set to current time")
	}
	if now.Sub(user.UpdatedAt) > time.Second {
		t.Error("UpdatedAt should be set to current time")
	}
}

func TestUser_BeforeUpdate(t *testing.T) {
	oldTime := time.Now().Add(-1 * time.Hour)
	user := &User{
		Name:      "Test User",
		Email:     "test@example.com",
		CreatedAt: oldTime,
		UpdatedAt: oldTime,
	}

	// BeforeUpdate should update UpdatedAt only
	err := user.BeforeUpdate(nil)
	if err != nil {
		t.Errorf("BeforeUpdate() returned unexpected error: %v", err)
	}

	if user.UpdatedAt.IsZero() {
		t.Error("BeforeUpdate() should set UpdatedAt, but it's zero")
	}

	// Verify UpdatedAt is recent (within 1 second)
	now := time.Now()
	if now.Sub(user.UpdatedAt) > time.Second {
		t.Error("UpdatedAt should be updated to current time")
	}

	// Verify CreatedAt was not changed
	if user.CreatedAt != oldTime {
		t.Error("BeforeUpdate() should not modify CreatedAt")
	}
}
