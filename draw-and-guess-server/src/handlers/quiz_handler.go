package handlers

import (
	"draw-and-guess-server/src/models"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetOXQuiz returns a random OX quiz
func GetOXQuiz(c *gin.Context) {
	quizzes := models.GetTestOXQuizzes()
	if len(quizzes) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "No OX quizzes available",
		})
		return
	}

	// Select random quiz
	quiz := quizzes[rand.Intn(len(quizzes))]

	c.JSON(http.StatusOK, gin.H{
		"result": quiz,
	})
}

// GetGeneralQuiz returns a random general quiz
func GetGeneralQuiz(c *gin.Context) {
	quizzes := models.GetTestGeneralQuizzes()
	if len(quizzes) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "No general quizzes available",
		})
		return
	}

	// Select random quiz
	quiz := quizzes[rand.Intn(len(quizzes))]

	c.JSON(http.StatusOK, gin.H{
		"result": quiz,
	})
}

// GetQuizByType returns a quiz based on game type
func GetQuizByType(c *gin.Context) {
	gameType := c.Query("type")
	if gameType == "" {
		gameType = c.Param("type")
	}

	switch gameType {
	case string(models.GameTypeOX):
		GetOXQuiz(c)
	case string(models.GameTypeGeneral):
		GetGeneralQuiz(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid game type. Must be 'ox' or 'general'",
		})
	}
}
