package handlers

import (
	"context"
	"draw-and-guess-server/src/database"
	"draw-and-guess-server/src/models"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetUser는 사용자 정보를 조회하는 API 핸들러
// GET /users/:id
// 용도: ID로 특정 사용자의 정보를 조회
func GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		JSONBadRequest(c, "user ID is required")
		return
	}

	// 문자열 ID를 ObjectID로 변환
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		JSONBadRequest(c, "invalid user ID format")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// MongoDB에서 사용자 조회
	var user models.User
	collection := database.GetCollection("users")
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		log.Printf("User %s not found: %v", userID, err)
		JSONNotFound(c, "user not found")
		return
	}

	JSONSuccess(c, user)
}

// CreateUser는 새로운 사용자를 생성하는 API 핸들러
// POST /users
// Form 데이터: nickname, password, birth_date (YYYY-MM-DD 형식)
// 용도: 새 사용자 계정을 생성하고 DB에 저장
func CreateUser(c *gin.Context) {
	nickname := c.PostForm("nickname")
	password := c.PostForm("password")
	birthDate := c.PostForm("birth_date")

	// 닉네임 길이 검증 (3~20자)
	if nicknameLen := len(nickname); nicknameLen < 3 || nicknameLen > 20 {
		JSONBadRequest(c, "nickname must be between 3 and 20 characters")
		return
	}

	// 비밀번호 필수 확인
	if password == "" {
		JSONBadRequest(c, "password is missing")
		return
	}

	// 생년월일 파싱 및 검증
	parsedBirthDate, err := time.Parse("2006-01-02", birthDate)
	if err != nil {
		JSONBadRequest(c, "invalid birth_date format, expected YYYY-MM-DD")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := database.GetCollection("users")

	// 닉네임 중복 확인
	count, err := collection.CountDocuments(ctx, bson.M{"nickname": nickname})
	if err != nil {
		log.Printf("Failed to check nickname existence: %v", err)
		JSONInternalError(c, "database error")
		return
	}
	if count > 0 {
		JSONConflict(c, "nickname already exists")
		return
	}

	// 새 사용자 생성
	user := models.User{
		Nickname:     nickname,
		Password:     password,
		BirthDate:    parsedBirthDate,
		Money:        0,
		ProfileImage: "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// MongoDB에 삽입
	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("Failed to create user %s: %v", nickname, err)
		JSONInternalError(c, "failed to create user")
		return
	}

	log.Printf("Created user: %s", nickname)

	JSONSuccess(c, map[string]interface{}{
		"id":       result.InsertedID,
		"nickname": nickname,
	})
}

// UpdateUser는 사용자 정보를 수정하는 API 핸들러
// PATCH /users/:id
// Form 데이터: password (선택), birth_date (선택), profile_image (선택)
// 용도: 사용자의 정보를 부분적으로 업데이트
func UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		JSONBadRequest(c, "user ID is required")
		return
	}

	// 문자열 ID를 ObjectID로 변환
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		JSONBadRequest(c, "invalid user ID format")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := database.GetCollection("users")

	// 업데이트 문서 작성
	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	// 제공된 필드만 업데이트
	if password := c.PostForm("password"); password != "" {
		update["$set"].(bson.M)["password"] = password
	}
	if birthDate := c.PostForm("birth_date"); birthDate != "" {
		parsedBirthDate, err := time.Parse("2006-01-02", birthDate)
		if err != nil {
			JSONBadRequest(c, "invalid birth_date format, expected YYYY-MM-DD")
			return
		}
		update["$set"].(bson.M)["birth_date"] = parsedBirthDate
	}
	if profileImage := c.PostForm("profile_image"); profileImage != "" {
		update["$set"].(bson.M)["profile_image"] = profileImage
	}

	// MongoDB 업데이트 실행
	result, err := collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		log.Printf("Failed to update user %s: %v", userID, err)
		JSONInternalError(c, "failed to update user")
		return
	}

	if result.MatchedCount == 0 {
		JSONNotFound(c, "user not found")
		return
	}

	log.Printf("Updated user: %s", userID)

	JSONSuccess(c, map[string]interface{}{
		"id":             userID,
		"modified_count": result.ModifiedCount,
	})
}
