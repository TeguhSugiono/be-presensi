package response

import "github.com/gin-gonic/gin"

func Success(c *gin.Context, status int, message string, data interface{}) {
	if data == nil {
		data = []interface{}{}
	}

	c.JSON(status, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"message": message,
		"data":    []interface{}{},
	})
}
