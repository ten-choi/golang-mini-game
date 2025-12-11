package handlers

import (
	"draw-and-guess-server/models"
	"draw-and-guess-server/valkey"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetUser(roomID string) {
	for range ticker.C {
		var room models.GameRoom
		if err := valkey.GetJSON(roomKeyPrefix+roomID, &room); err != nil {
			log.Printf("Timer: Room %s not found, stopping timer", roomID)
			return
		}

	}
}

func CreateUser(c *gin.Context) {
	nickname := c.PostForm("nickname")
	password := c.PostForm("password")
	birthDate := c.PostForm("birth_date")

	if nickname == "" {
		respondError(c, http.StatusBadRequest, "nickname is missing")
		return
	}
	if password == "" {
		respondError(c, http.StatusBadRequest, "password is missing")
		return
	}
	parsedBirthDate, err := time.Parse("2006-01-02", birthDate)

	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid birth_date format, expected YYYY-MM-DD")
		user := models.User{
			Nickname:     nickname,
			PassWrod:     password,
			BirthDate:    parsedBirthDate,
			WinningPoint: 0,
			ProfileImage: "",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		append(user)

		log.Printf("Created user : %s", nickname)

		respondSuccess(c, "Insert Game Room", map[string]interface{}{
			"room_id": nickname,
		})
	}

}
