package graph

import (
	"context"
	"draw-and-guess-server/internal/graph/model"
	"draw-and-guess-server/internal/service"
	"fmt"
)

// ========================================
// User Mutations
// ========================================

// CreateUser is the resolver for the createUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
	// GraphQL input -> Service DTO
	dto := service.CreateUserDTO{
		HangeID: input.HangeID,
		Name: func() string {
			if input.Name == nil {
				return input.HangeID
			}
			return *input.Name
		}(),
		AvatarURL: "",
		Level:     0,
		Credit:    0,
	}
	if input.AvatarURL != nil {
		dto.AvatarURL = *input.AvatarURL
	}

	created, err := r.UserService.CreateUser(ctx, dto)
	if err != nil {
		return nil, err
	}

	// Domain model -> GraphQL model
	return &model.User{
		ID:        created.ID.Hex(),
		HangeID:   created.HangeID,
		Name:      created.Name,
		AvatarURL: &created.AvatarURL,
		Level:     int32(created.Level),
		Credit:    int32(created.Credit),
		CreatedAt: created.CreatedAt,
		UpdatedAt: created.UpdatedAt,
	}, nil
}

// UpdateUser is the resolver for the updateUser field.
func (r *mutationResolver) UpdateUser(ctx context.Context, name string, input model.UpdateUserInput) (*model.User, error) {
	// GraphQL input -> Service DTO
	dto := service.UpdateUserDTO{
		AvatarURL: input.AvatarURL,
	}
	if input.Level != nil {
		level := int(*input.Level)
		dto.Level = &level
	}
	if input.Credit != nil {
		credit := int(*input.Credit)
		dto.Credit = &credit
	}

	updated, err := r.UserService.UpdateUser(ctx, name, dto)
	if err != nil {
		return nil, err
	}

	// Domain model -> GraphQL model
	return &model.User{
		ID:        updated.ID.Hex(),
		HangeID:   updated.HangeID,
		Name:      updated.Name,
		AvatarURL: &updated.AvatarURL,
		Level:     int32(updated.Level),
		Credit:    int32(updated.Credit),
		CreatedAt: updated.CreatedAt,
		UpdatedAt: updated.UpdatedAt,
	}, nil
}

// ========================================
// User Queries
// ========================================

// UserByHangeID is the resolver for the userByHangeId field.
func (r *queryResolver) UserByHangeID(ctx context.Context, hangeId string) (*model.User, error) {
	user, err := r.UserService.GetUserByHangeID(ctx, hangeId)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:        user.ID.Hex(),
		HangeID:   user.HangeID,
		Name:      user.Name,
		AvatarURL: &user.AvatarURL,
		Level:     int32(user.Level),
		Credit:    int32(user.Credit),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// User is the resolver for the user field.
func (r *queryResolver) UserByName(ctx context.Context, Name string) (*model.User, error) {
	user, err := r.UserService.GetUserByName(ctx, Name)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:        user.ID.Hex(),
		HangeID:   user.HangeID,
		Name:      user.Name,
		AvatarURL: &user.AvatarURL,
		Level:     int32(user.Level),
		Credit:    int32(user.Credit),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context, limit *int32, offset *int32, search *string, minLevel *int32) ([]*model.User, error) {
	users, err := r.UserService.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.User, 0)
	for _, u := range users {
		// Apply filters
		if minLevel != nil && int32(u.Level) < *minLevel {
			continue
		}
		if search != nil && len(*search) > 0 {
			// Simple case-insensitive search in name
			if !contains(u.Name, *search) {
				continue
			}
		}

		result = append(result, &model.User{
			ID:        u.ID.Hex(),
			HangeID:   u.HangeID,
			Name:      u.Name,
			AvatarURL: &u.AvatarURL,
			Level:     int32(u.Level),
			Credit:    int32(u.Credit),
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		})

		// Apply limit
		if limit != nil && len(result) >= int(*limit) {
			break
		}
	}

	// Apply offset
	if offset != nil && int(*offset) < len(result) {
		result = result[*offset:]
	}

	return result, nil
}

// Helper function for case-insensitive string contains
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr ||
		len(substr) == 0 ||
		indexIgnoreCase(str, substr) >= 0)
}

func indexIgnoreCase(str, substr string) int {
	str = toLower(str)
	substr = toLower(substr)
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// DeleteUser is the resolver for the deleteUser field.
func (r *mutationResolver) DeleteUser(ctx context.Context, Name string) (bool, error) {
	// User deletion is not supported for data integrity
	// Consider deactivation or soft delete instead
	return false, fmt.Errorf("user deletion is not supported")
}
