package repository

import (
	"context"
	"time"

	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	GetByNickname(ctx context.Context, nickname string) (*models.User, error)
	GetAll(ctx context.Context) ([]*models.User, error)
	Create(ctx context.Context, user *models.User) (*models.User, error)
	Update(ctx context.Context, nickname string, avatarURL *string, level, credit *int) (*models.User, error)
}

type userRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new user repository
func NewUserRepository(collection *mongo.Collection) UserRepository {
	return &userRepository{collection: collection}
}

func (r *userRepository) GetByNickname(ctx context.Context, nickname string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"nickname": nickname}).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return nil, common.NewNotFoundError("user not found")
	}
	if err != nil {
		return nil, common.NewInternalError("failed to get user", err)
	}

	return &user, nil
}

func (r *userRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, common.NewInternalError("failed to query users", err)
	}
	defer cursor.Close(ctx)

	var users []*models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, common.NewInternalError("failed to decode users", err)
	}

	return users, nil
}

func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	// Generate MongoDB ObjectID
	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, common.NewInternalError("nickname already exists", err)
		}
		return nil, common.NewInternalError("failed to create user", err)
	}

	return user, nil
}

func (r *userRepository) Update(ctx context.Context, nickname string, avatarURL *string, level, credit *int) (*models.User, error) {
	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	setFields := update["$set"].(bson.M)
	if avatarURL != nil {
		setFields["avatar_url"] = *avatarURL
	}
	if level != nil {
		setFields["level"] = *level
	}
	if credit != nil {
		setFields["credit"] = *credit
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var user models.User
	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"nickname": nickname},
		update,
		opts,
	).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return nil, common.NewNotFoundError("user not found")
	}
	if err != nil {
		return nil, common.NewInternalError("failed to update user", err)
	}

	return &user, nil
}
