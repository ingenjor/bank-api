package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"bank-api/internal/models"
	"bank-api/internal/service"
)

type mockAuthRepo struct {
	mock.Mock
}

func (m *mockAuthRepo) Create(ctx context.Context, u *models.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}
func (m *mockAuthRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if u, ok := args.Get(0).(*models.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(0)
}
func (m *mockAuthRepo) IsUnique(ctx context.Context, email, username string) (bool, error) {
	args := m.Called(ctx, email, username)
	return args.Bool(0), args.Error(1)
}

func TestAuthService_Register_Success(t *testing.T) {
	repo := new(mockAuthRepo)
	svc := service.NewAuthService(repo, "secret")
	req := models.RegisterRequest{Username: "user1", Email: "user1@x.com", Password: "password123"}
	repo.On("IsUnique", mock.Anything, req.Email, req.Username).Return(true, nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
	err := svc.Register(context.Background(), req)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAuthService_Register_Duplicate(t *testing.T) {
	repo := new(mockAuthRepo)
	svc := service.NewAuthService(repo, "secret")
	req := models.RegisterRequest{Username: "dup", Email: "dup@x.com", Password: "password123"}
	repo.On("IsUnique", mock.Anything, req.Email, req.Username).Return(false, nil)
	err := svc.Register(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already taken")
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := new(mockAuthRepo)
	svc := service.NewAuthService(repo, "secret")
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.DefaultCost)
	user := &models.User{ID: "uid", PasswordHash: string(hash)}
	repo.On("GetByEmail", mock.Anything, "a@x.com").Return(user, nil)
	token, err := svc.Login(context.Background(), models.LoginRequest{Email: "a@x.com", Password: "pass123"})
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	repo := new(mockAuthRepo)
	svc := service.NewAuthService(repo, "secret")
	user := &models.User{ID: "uid", PasswordHash: "$2a$10$..."}
	repo.On("GetByEmail", mock.Anything, "a@x.com").Return(user, nil)
	_, err := svc.Login(context.Background(), models.LoginRequest{Email: "a@x.com", Password: "wrong"})
	assert.Error(t, err)
}
