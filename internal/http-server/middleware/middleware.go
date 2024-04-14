package middleware

import "github.com/gin-gonic/gin"

func GetCreator() gin.HandlerFunc {
	return func(c *gin.Context) {
		creator, _, ok := c.Request.BasicAuth()
		if ok {
			c.Set("username", creator)
		}

		c.Next()
	}
}
