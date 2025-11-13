package repositories

import (
	"testing"

	"velure-auth-service/internal/models"
	"velure-auth-service/internal/testutil"

	"gorm.io/gorm"
)

func TestPasswordResetRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewPasswordResetRepository(db)
	userRepo := NewUserRepository(db)

	user := testutil.CreateTestUser()
	user.ID = 0
	userRepo.Create(user)

	reset := testutil.CreateTestPasswordReset(user.ID)
	reset.ID = 0

	err := repo.Create(reset)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	if reset.ID == 0 {
		t.Error("Create() should set password reset ID but it remains 0")
	}
}

func TestPasswordResetRepository_GetByToken(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewPasswordResetRepository(db)
	userRepo := NewUserRepository(db)

	user := testutil.CreateTestUser()
	user.ID = 0
	userRepo.Create(user)

	reset := testutil.CreateTestPasswordReset(user.ID)
	reset.ID = 0
	reset.Token = "unique-reset-token-456"
	repo.Create(reset)

	found, err := repo.GetByToken("unique-reset-token-456")
	if err != nil {
		t.Fatalf("GetByToken() error = %v", err)
	}

	if found.Token != "unique-reset-token-456" {
		t.Errorf("GetByToken() token = %s, want unique-reset-token-456", found.Token)
	}

	// Test not found
	_, err = repo.GetByToken("non-existent-token")
	if err == nil {
		t.Error("GetByToken() should return error for non-existent token")
	}
}

func TestPasswordResetRepository_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewPasswordResetRepository(db)
	userRepo := NewUserRepository(db)

	user := testutil.CreateTestUser()
	user.ID = 0
	userRepo.Create(user)

	reset := testutil.CreateTestPasswordReset(user.ID)
	reset.ID = 0
	repo.Create(reset)

	resetID := reset.ID

	err := repo.Delete(resetID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	var deleted models.PasswordReset
	result := db.First(&deleted, resetID)
	if result.Error != gorm.ErrRecordNotFound {
		t.Error("Delete() password reset still exists after deletion")
	}
}
