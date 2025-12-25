package mocks

import (
	"draw-and-guess-server/internal/models"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	Users map[string]*models.User
}

// NewMockUserRepository creates a new mock user repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users: make(map[string]*models.User),
	}
}

// GetByUsername returns a user by username
func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	if user, exists := m.Users[username]; exists {
		return user, nil
	}
	return nil, nil
}

// GetAll returns all users
func (m *MockUserRepository) GetAll() ([]*models.User, error) {
	users := make([]*models.User, 0, len(m.Users))
	for _, user := range m.Users {
		users = append(users, user)
	}
	return users, nil
}

// Create creates a new user
func (m *MockUserRepository) Create(user *models.User) error {
	m.Users[user.Username] = user
	return nil
}

// Update updates a user
func (m *MockUserRepository) Update(user *models.User) error {
	m.Users[user.Username] = user
	return nil
}
