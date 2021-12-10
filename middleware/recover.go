package middleware

import "github.com/gin-gonic/gin"

func Recover(handler func(*gin.Context, interface{})) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				handler(c, err)
				c.Abort()
			}
		}()
		c.Next()
	}
}
