package test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"

	"boilerplate/services/auth/domain"
	"boilerplate/services/auth/usecase"
	postgresAdapter "boilerplate/shared/adapter/postgres_adapter"
)

const (
	testConnString = "postgres://postgres:root@localhost:5432/postgres_test?sslmode=disable"
	testJWTSecret  = "test-jwt-secret-key"
)

func setupTestDB(t *testing.T) (*sql.DB, domain.AuthUseCase) {
	t.Helper()

	db, err := sql.Open("postgres", testConnString)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("failed to clean users table: %v", err)
	}

	repo := postgresAdapter.NewAuthRepository(db, testJWTSecret)
	uc := usecase.NewAuthUseCase(repo)

	return db, uc
}

func TestSignup(t *testing.T) {
	db, uc := setupTestDB(t)
	defer db.Close()

	t.Run("Success", func(t *testing.T) {
		user, err := uc.Signup(context.Background(), "testuser", "password123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user == nil {
			t.Fatal("expected user, got nil")
		}
		if user.Username != "testuser" {
			t.Errorf("expected username 'testuser', got '%s'", user.Username)
		}
		if user.ID == "" {
			t.Error("expected non-empty user ID")
		}
	})

	t.Run("Duplicate Username", func(t *testing.T) {
		_, err := uc.Signup(context.Background(), "duplicateuser", "password123")
		if err != nil {
			t.Fatalf("first signup failed: %v", err)
		}

		_, err = uc.Signup(context.Background(), "duplicateuser", "password456")
		if err == nil {
			t.Fatal("expected error for duplicate username, got nil")
		}
	})
}

func TestLogin(t *testing.T) {
	db, uc := setupTestDB(t)
	defer db.Close()

	_, err := uc.Signup(context.Background(), "loginuser", "correctpassword")
	if err != nil {
		t.Fatalf("failed to signup test user: %v", err)
	}

	t.Run("Success", func(t *testing.T) {
		token, err := uc.Login(context.Background(), "loginuser", "correctpassword")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token == "" {
			t.Fatal("expected non-empty token")
		}
	})

	t.Run("Wrong Password", func(t *testing.T) {
		_, err := uc.Login(context.Background(), "loginuser", "wrongpassword")
		if err == nil {
			t.Fatal("expected error for wrong password, got nil")
		}
	})

	t.Run("Nonexistent User", func(t *testing.T) {
		_, err := uc.Login(context.Background(), "nobody", "password")
		if err == nil {
			t.Fatal("expected error for nonexistent user, got nil")
		}
	})
}

func TestValidateToken(t *testing.T) {
	db, uc := setupTestDB(t)
	defer db.Close()

	_, err := uc.Signup(context.Background(), "tokenuser", "password123")
	if err != nil {
		t.Fatalf("failed to signup test user: %v", err)
	}

	token, err := uc.Login(context.Background(), "tokenuser", "password123")
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	t.Run("Valid Token", func(t *testing.T) {
		user, err := uc.ValidateToken(context.Background(), token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user == nil {
			t.Fatal("expected user, got nil")
		}
		if user.Username != "tokenuser" {
			t.Errorf("expected username 'tokenuser', got '%s'", user.Username)
		}
	})

	t.Run("Invalid Token", func(t *testing.T) {
		_, err := uc.ValidateToken(context.Background(), "invalid-token-string")
		if err == nil {
			t.Fatal("expected error for invalid token, got nil")
		}
	})

	t.Run("Empty Token", func(t *testing.T) {
		_, err := uc.ValidateToken(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty token, got nil")
		}
	})
}

func TestFullAuthFlow(t *testing.T) {
	db, uc := setupTestDB(t)
	defer db.Close()

	t.Run("Signup -> Login -> Validate", func(t *testing.T) {
		user, err := uc.Signup(context.Background(), "flowuser", "secretpass")
		if err != nil {
			t.Fatalf("signup failed: %v", err)
		}

		token, err := uc.Login(context.Background(), "flowuser", "secretpass")
		if err != nil {
			t.Fatalf("login failed: %v", err)
		}

		validatedUser, err := uc.ValidateToken(context.Background(), token)
		if err != nil {
			t.Fatalf("token validation failed: %v", err)
		}

		if validatedUser.ID != user.ID {
			t.Errorf("expected user ID '%s', got '%s'", user.ID, validatedUser.ID)
		}
		if validatedUser.Username != user.Username {
			t.Errorf("expected username '%s', got '%s'", user.Username, validatedUser.Username)
		}
	})
}
