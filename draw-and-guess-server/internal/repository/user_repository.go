package repository

import (
	"context"
	"database/sql"

	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/models"
	"draw-and-guess-server/pkg/utils"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetAll(ctx context.Context) ([]*models.User, error)
	Create(ctx context.Context, user *models.User) (*models.User, error)
	Update(ctx context.Context, username string, displayName, email, avatarURL *string) (*models.User, error)
}

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, display_name, email, avatar_url, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Email,
		&user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, common.NewNotFoundError("user not found")
	}
	if err != nil {
		return nil, common.NewInternalError("failed to get user", err)
	}

	return user, nil
}

func (r *userRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, username, display_name, email, avatar_url, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, common.NewInternalError("failed to query users", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.DisplayName, &user.Email,
			&user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, common.NewInternalError("failed to scan user", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	// Generate Snowflake ID
	user.ID = utils.GenerateID()
	
	query := `
		INSERT INTO users (id, username, display_name, email, avatar_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		user.ID, user.Username, user.DisplayName, user.Email, user.AvatarURL,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, common.NewInternalError("failed to create user", err)
	}

	return user, nil
}

func (r *userRepository) Update(ctx context.Context, username string, displayName, email, avatarURL *string) (*models.User, error) {
	query := `
		UPDATE users
		SET display_name = COALESCE($2, display_name),
		    email = COALESCE($3, email),
		    avatar_url = COALESCE($4, avatar_url),
		    updated_at = CURRENT_TIMESTAMP
		WHERE username = $1
		RETURNING id, username, display_name, email, avatar_url, created_at, updated_at
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, username, displayName, email, avatarURL).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Email,
		&user.AvatarURL, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, common.NewNotFoundError("user not found")
	}
	if err != nil {
		return nil, common.NewInternalError("failed to update user", err)
	}

	return user, nil
}
