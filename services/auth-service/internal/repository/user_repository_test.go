package repositories

import (
	"context"
	"testing"

	"velure-auth-service/internal/model"
	"velure-auth-service/internal/testutil"

	"gorm.io/gorm"
)

func TestUserRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
	}{
		{
			name:    "successful user creation",
			user:    testutil.CreateTestUser(),
			wantErr: false,
		},
		{
			name: "create user with custom data",
			user: testutil.CreateTestUser(func(u *models.User) {
				u.Email = "custom@example.com"
				u.Name = "Custom User"
			}),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.SetupTestDB(t)
			repo := NewUserRepository(db)

			// Reset ID to 0 for auto-increment
			tt.user.ID = 0

			err := repo.Create(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.user.ID == 0 {
				t.Error("Create() should set user ID but it remains 0")
			}
		})
	}
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewUserRepository(db)

	user1 := testutil.CreateTestUser()
	user1.ID = 0
	err := repo.Create(user1)
	if err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	// Try to create another user with same email
	user2 := testutil.CreateTestUser()
	user2.ID = 0
	err = repo.Create(user2)
	if err == nil {
		t.Error("Create() should fail with duplicate email but succeeded")
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewUserRepository(db)

	// Create a test user
	testUser := testutil.CreateTestUser()
	testUser.ID = 0
	err := repo.Create(testUser)
	if err != nil {
		t.Fatalf("failed to setup test user: %v", err)
	}

	tests := []struct {
		name      string
		email     string
		wantUser  bool
		wantErr   bool
		checkName string
	}{
		{
			name:      "found existing user",
			email:     testUser.Email,
			wantUser:  true,
			wantErr:   false,
			checkName: testUser.Name,
		},
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			wantUser: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantUser {
				if user == nil {
					t.Error("GetByEmail() returned nil user when user was expected")
					return
				}
				if user.Email != tt.email {
					t.Errorf("GetByEmail() email = %v, want %v", user.Email, tt.email)
				}
				if user.Name != tt.checkName {
					t.Errorf("GetByEmail() name = %v, want %v", user.Name, tt.checkName)
				}
			} else {
				if user != nil {
					t.Error("GetByEmail() returned user when none was expected")
				}
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewUserRepository(db)

	// Create a test user
	testUser := testutil.CreateTestUser()
	testUser.ID = 0
	err := repo.Create(testUser)
	if err != nil {
		t.Fatalf("failed to setup test user: %v", err)
	}

	tests := []struct {
		name     string
		id       uint
		wantUser bool
		wantErr  bool
	}{
		{
			name:     "found existing user",
			id:       testUser.ID,
			wantUser: true,
			wantErr:  false,
		},
		{
			name:     "user not found",
			id:       999,
			wantUser: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := repo.GetByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantUser {
				if user == nil {
					t.Error("GetByID() returned nil user when user was expected")
					return
				}
				if user.ID != tt.id {
					t.Errorf("GetByID() id = %v, want %v", user.ID, tt.id)
				}
			} else {
				if user != nil {
					t.Error("GetByID() returned user when none was expected")
				}
			}
		})
	}
}

func TestUserRepository_GetAll(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewUserRepository(db)

	// Create multiple test users
	users := testutil.CreateTestUsers(3)
	for _, user := range users {
		user.ID = 0
		err := repo.Create(user)
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}
	}

	allUsers, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(allUsers) != 3 {
		t.Errorf("GetAll() returned %d users, want 3", len(allUsers))
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewUserRepository(db)

	// Create a test user
	testUser := testutil.CreateTestUser()
	testUser.ID = 0
	err := repo.Create(testUser)
	if err != nil {
		t.Fatalf("failed to setup test user: %v", err)
	}

	// Update user
	testUser.Name = "Updated Name"
	testUser.Email = "updated@example.com"

	err = repo.Update(testUser)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update
	updated, err := repo.GetByID(testUser.ID)
	if err != nil {
		t.Fatalf("failed to get updated user: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Update() name = %v, want 'Updated Name'", updated.Name)
	}
	if updated.Email != "updated@example.com" {
		t.Errorf("Update() email = %v, want 'updated@example.com'", updated.Email)
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewUserRepository(db)

	// Create a test user
	testUser := testutil.CreateTestUser()
	testUser.ID = 0
	err := repo.Create(testUser)
	if err != nil {
		t.Fatalf("failed to setup test user: %v", err)
	}

	userID := testUser.ID

	// Delete user
	err = repo.Delete(userID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(userID)
	if err == nil {
		t.Error("Delete() user still exists after deletion")
	}
	if err != gorm.ErrRecordNotFound {
		t.Errorf("Delete() expected ErrRecordNotFound, got %v", err)
	}
}

func TestUserRepository_GetByPage(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewUserRepository(db)

	// Create 10 test users
	for i := 0; i < 10; i++ {
		user := testutil.CreateTestUser(func(u *models.User) {
			u.ID = 0
			u.Email = u.Email + string(rune('0'+i))
		})
		err := repo.Create(user)
		if err != nil {
			t.Fatalf("failed to create test user: %v", err)
		}
	}

	tests := []struct {
		name      string
		page      int
		pageSize  int
		wantCount int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:      "first page with 5 items",
			page:      1,
			pageSize:  5,
			wantCount: 5,
			wantTotal: 10,
			wantErr:   false,
		},
		{
			name:      "second page with 5 items",
			page:      2,
			pageSize:  5,
			wantCount: 5,
			wantTotal: 10,
			wantErr:   false,
		},
		{
			name:      "page with 3 items",
			page:      1,
			pageSize:  3,
			wantCount: 3,
			wantTotal: 10,
			wantErr:   false,
		},
		{
			name:      "last page partially filled",
			page:      4,
			pageSize:  3,
			wantCount: 1,
			wantTotal: 10,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, total, err := repo.GetByPage(tt.page, tt.pageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByPage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(users) != tt.wantCount {
				t.Errorf("GetByPage() returned %d users, want %d", len(users), tt.wantCount)
			}

			if total != tt.wantTotal {
				t.Errorf("GetByPage() total = %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

func TestUserRepository_CountUsers(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewUserRepository(db)

	user1 := testutil.CreateTestUser(func(u *models.User) {
		u.ID = 0
		u.Email = "count1@example.com"
	})
	user2 := testutil.CreateTestUser(func(u *models.User) {
		u.ID = 0
		u.Email = "count2@example.com"
	})

	if err := repo.Create(user1); err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}
	if err := repo.Create(user2); err != nil {
		t.Fatalf("failed to create second user: %v", err)
	}

	count, err := repo.CountUsers(context.Background())
	if err != nil {
		t.Fatalf("CountUsers() error = %v", err)
	}
	if count != 2 {
		t.Fatalf("CountUsers() = %d, want 2", count)
	}

	count, err = repo.CountUsers(nil)
	if err != nil {
		t.Fatalf("CountUsers(nil) error = %v", err)
	}
	if count != 2 {
		t.Fatalf("CountUsers(nil) = %d, want 2", count)
	}
}

func TestUserRepository_CountUsers_Error(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewUserRepository(db)

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql db: %v", err)
	}
	sqlDB.Close()

	if _, err := repo.CountUsers(context.Background()); err == nil {
		t.Fatalf("expected error when counting with closed DB")
	}
}

func TestUserRepository_GetByPage_Error(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewUserRepository(db)

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql db: %v", err)
	}
	sqlDB.Close()

	if _, _, err := repo.GetByPage(1, 10); err == nil {
		t.Fatalf("expected error when querying with closed DB")
	}
}
