package auth

import (
	"fmt"

	"github.com/ChenKS12138/remote-terminal/dao"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func RequireAuth(permission PermissionType) gin.HandlerFunc {
	configDao := dao.NewConfigDaoMust()
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie(COOKIE_NAME)
		if err != nil {
			panic(err)
		}
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return configDao.JwtSecret, nil
		})
		if err != nil || !token.Valid {
			panic(err)
		}
		c.Keys[GIN_CLAIM_KEY] = token.Claims
		c.Next()
		delete(c.Keys, GIN_CLAIM_KEY)
	}
}
