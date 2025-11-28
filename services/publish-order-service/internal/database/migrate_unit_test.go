package database

import (
	"database/sql"
	"errors"
	"testing"
)

type stubMigrator struct {
	version uint
	dirty   bool
	upErr   error
}

func (s *stubMigrator) Version() (uint, bool, error) {
	return s.version, s.dirty, nil
}

func (s *stubMigrator) Up() error {
	return s.upErr
}

func TestRunMigrations_NoChange(t *testing.T) {
	orig := newMigrator
	defer func() { newMigrator = orig }()

	stub := &stubMigrator{version: 1}
	newMigrator = func(db *sql.DB, path string) (migrator, error) {
		return stub, nil
	}

	if err := RunMigrations(&sql.DB{}, "/tmp"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRunMigrations_UpError(t *testing.T) {
	orig := newMigrator
	defer func() { newMigrator = orig }()

	stub := &stubMigrator{upErr: errors.New("boom")}
	newMigrator = func(db *sql.DB, path string) (migrator, error) {
		return stub, nil
	}

	err := RunMigrations(&sql.DB{}, "/tmp")
	if err == nil {
		t.Fatal("expected error on Up failure")
	}
	if err.Error() == "" {
		t.Fatal("expected error message")
	}
}
