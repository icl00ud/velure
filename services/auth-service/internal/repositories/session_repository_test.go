package repositories

import (
	"context"
	"testing"
	"time"

	"velure-auth-service/internal/models"
	"velure-auth-service/internal/testutil"

	"gorm.io/gorm"
)

func TestSessionRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewSessionRepository(db)

	// Create user first
	userRepo := NewUserRepository(db)
	user := testutil.CreateTestUser()
	user.ID = 0
	userRepo.Create(user)

	session := testutil.CreateTestSession(user.ID)
	session.ID = 0

	err := repo.Create(session)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	if session.ID == 0 {
		t.Error("Create() should set session ID but it remains 0")
	}
}

func TestSessionRepository_GetByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewSessionRepository(db)
	userRepo := NewUserRepository(db)

	user := testutil.CreateTestUser()
	user.ID = 0
	userRepo.Create(user)

	session := testutil.CreateTestSession(user.ID)
	session.ID = 0
	repo.Create(session)

	found, err := repo.GetByUserID(user.ID)
	if err != nil {
		t.Fatalf("GetByUserID() error = %v", err)
	}

	if found.UserID != user.ID {
		t.Errorf("GetByUserID() userID = %d, want %d", found.UserID, user.ID)
	}

	// Test not found
	_, err = repo.GetByUserID(999)
	if err == nil {
		t.Error("GetByUserID() should return error for non-existent user")
	}
}

func TestSessionRepository_GetByRefreshToken(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewSessionRepository(db)
	userRepo := NewUserRepository(db)

	user := testutil.CreateTestUser()
	user.ID = 0
	userRepo.Create(user)

	session := testutil.CreateTestSession(user.ID)
	session.ID = 0
	session.RefreshToken = "unique-refresh-token-123"
	repo.Create(session)

	found, err := repo.GetByRefreshToken("unique-refresh-token-123")
	if err != nil {
		t.Fatalf("GetByRefreshToken() error = %v", err)
	}

	if found.RefreshToken != "unique-refresh-token-123" {
		t.Errorf("GetByRefreshToken() token = %s, want unique-refresh-token-123", found.RefreshToken)
	}

	// Test not found
	_, err = repo.GetByRefreshToken("non-existent-token")
	if err == nil {
		t.Error("GetByRefreshToken() should return error for non-existent token")
	}
}

func TestSessionRepository_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewSessionRepository(db)
	userRepo := NewUserRepository(db)

	user := testutil.CreateTestUser()
	user.ID = 0
	userRepo.Create(user)

	session := testutil.CreateTestSession(user.ID)
	session.ID = 0
	repo.Create(session)

	session.AccessToken = "new-access-token"
	session.RefreshToken = "new-refresh-token"

	err := repo.Update(session)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	updated, _ := repo.GetByUserID(user.ID)
	if updated.AccessToken != "new-access-token" {
		t.Errorf("Update() accessToken = %s, want new-access-token", updated.AccessToken)
	}
}

func TestSessionRepository_InvalidateByRefreshToken(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewSessionRepository(db)
	userRepo := NewUserRepository(db)

	user := testutil.CreateTestUser()
	user.ID = 0
	userRepo.Create(user)

	session := testutil.CreateTestSession(user.ID)
	session.ID = 0
	session.RefreshToken = "token-to-invalidate"
	session.ExpiresAt = time.Now().Add(24 * time.Hour)
	repo.Create(session)

	err := repo.InvalidateByRefreshToken("token-to-invalidate")
	if err != nil {
		t.Fatalf("InvalidateByRefreshToken() error = %v", err)
	}

	invalidated, _ := repo.GetByRefreshToken("token-to-invalidate")
	if invalidated.ExpiresAt.After(time.Now()) {
		t.Error("InvalidateByRefreshToken() session should be expired")
	}
}

func TestSessionRepository_CountActiveSessions(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewSessionRepository(db)
	userRepo := NewUserRepository(db)

	user := testutil.CreateTestUser()
	user.ID = 0
	userRepo.Create(user)

	activeSession := testutil.CreateTestSession(user.ID)
	activeSession.ID = 0
	activeSession.ExpiresAt = time.Now().Add(24 * time.Hour)
	repo.Create(activeSession)

	expiredSession := testutil.CreateTestSession(user.ID)
	expiredSession.ID = 0
	expiredSession.ExpiresAt = time.Now().Add(-1 * time.Hour)
	repo.Create(expiredSession)

	count, err := repo.CountActiveSessions(context.Background())
	if err != nil {
		t.Fatalf("CountActiveSessions() error = %v", err)
	}

	if count != 1 {
		t.Errorf("CountActiveSessions() = %d, want 1", count)
	}
}

func TestSessionRepository_CountActiveSessions_NilContext(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewSessionRepository(db)
	userRepo := NewUserRepository(db)

	user := testutil.CreateTestUser()
	user.ID = 0
	userRepo.Create(user)

	session := testutil.CreateTestSession(user.ID, func(s *models.Session) {
		s.ID = 0
		s.ExpiresAt = time.Now().Add(1 * time.Hour)
	})
	repo.Create(session)

	count, err := repo.CountActiveSessions(nil)
	if err != nil {
		t.Fatalf("CountActiveSessions(nil) error = %v", err)
	}
	if count != 1 {
		t.Fatalf("CountActiveSessions(nil) = %d, want 1", count)
	}
}

func TestSessionRepository_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewSessionRepository(db)
	userRepo := NewUserRepository(db)

	user := testutil.CreateTestUser()
	user.ID = 0
	userRepo.Create(user)

	session := testutil.CreateTestSession(user.ID)
	session.ID = 0
	repo.Create(session)

	sessionID := session.ID

	err := repo.Delete(sessionID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	var deleted models.Session
	result := db.First(&deleted, sessionID)
	if result.Error != gorm.ErrRecordNotFound {
		t.Error("Delete() session still exists after deletion")
	}
}
