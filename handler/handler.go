package handler

import (
	"github.com/labstack/echo"
)

// InitRouting ルーティング
func InitRouting(e *echo.Echo) {
	// github.io group
	githubIo := e.Group("/githubio")
	githubIo.POST("/contact", contact)
	githubIo.GET("/qanda", qAnda)
	githubIo.POST("/qanda/q", qAndaQCreate)
	githubIo.POST("/qanda/:question_id/good", qAndaQuestionGoodCount)
	githubIo.GET("/qanda/cookie", qAndaCookie)
	githubIo.POST("/qanda/cookie/accept", qAndaCookieAccept)
}
