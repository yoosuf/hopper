package users

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// MockRepository is a mock implementation of the Repository interface for testing
type MockRepository struct {
	users map[uuid.UUID]*User
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		users: make(map[uuid.UUID]*User),
	}
}

func (m *MockRepository) Create(ctx context.Context, user *User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return m.users[id], nil
}

func (m *MockRepository) Update(ctx context.Context, user *User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.users, id)
	return nil
}

func (m *MockRepository) List(ctx context.Context) ([]*User, error) {
	var users []*User
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func TestService_CreateUser(t *testing.T) {
	mockRepo := NewMockRepository()
	service := New(mockRepo)

	req := &CreateUserRequest{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
	}

	user, err := service.CreateUser(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, user.Email)
	}

	if user.Role != req.Role {
		t.Errorf("Expected role %s, got %s", req.Role, user.Role)
	}
}

func TestService_GetUser(t *testing.T) {
	mockRepo := NewMockRepository()
	service := New(mockRepo)

	userID := uuid.New()
	testUser := &User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
	}

	mockRepo.users[userID] = testUser

	user, err := service.GetUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if user.ID != userID {
		t.Errorf("Expected ID %s, got %s", userID, user.ID)
	}
}
