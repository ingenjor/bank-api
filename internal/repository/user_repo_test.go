package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"bank-api/internal/models"
	"bank-api/internal/repository"
)

func TestUserRepo_Create(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := repository.NewUserRepo(db)
	u := &models.User{ID: "id1", Username: "u", Email: "u@x.com", PasswordHash: "hash"}
	mock.ExpectExec("INSERT INTO users").
		WithArgs("id1", "u", "u@x.com", "hash").
		WillReturnResult(sqlmock.NewResult(1, 1))
	err := repo.Create(context.Background(), u)
	assert.NoError(t, err)
}

func TestUserRepo_GetByEmail_Found(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := repository.NewUserRepo(db)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "created_at"}).
		AddRow("id1", "user1", "a@b.com", "$2a$...", now)
	mock.ExpectQuery("SELECT id, username, email, password_hash, created_at FROM users WHERE email=\\$1").
		WithArgs("a@b.com").WillReturnRows(rows)
	user, err := repo.GetByEmail(context.Background(), "a@b.com")
	assert.NoError(t, err)
	assert.Equal(t, "user1", user.Username)
}

func TestUserRepo_GetByEmail_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := repository.NewUserRepo(db)
	mock.ExpectQuery("SELECT .* FROM users WHERE email=\\$1").
		WithArgs("no@x.com").WillReturnError(sql.ErrNoRows)
	user, err := repo.GetByEmail(context.Background(), "no@x.com")
	assert.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserRepo_GetByID(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := repository.NewUserRepo(db)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "created_at"}).
		AddRow("id1", "user1", "a@b.com", "hash", now)
	mock.ExpectQuery("SELECT .* FROM users WHERE id=\\$1").WithArgs("id1").WillReturnRows(rows)
	user, err := repo.GetByID(context.Background(), "id1")
	assert.NoError(t, err)
	assert.Equal(t, "user1", user.Username)
}

func TestUserRepo_IsUnique(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := repository.NewUserRepo(db)
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users WHERE email=\\$1 OR username=\\$2\\)").
		WithArgs("unique@x.com", "unique").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	unique, err := repo.IsUnique(context.Background(), "unique@x.com", "unique")
	assert.NoError(t, err)
	assert.True(t, unique)
}
