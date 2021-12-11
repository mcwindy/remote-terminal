package auth

import (
	"errors"

	"github.com/ChenKS12138/remote-terminal/dao"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var ErrAuthorizationFail = errors.New("authorization fail")

func RequireAuth(permission PermissionType) gin.HandlerFunc {
	return func(c *gin.Context) {
		configDao := dao.NewConfigDaoMust()
		tokenStr, err := c.Cookie(COOKIE_NAME)
		if err != nil {
			panic(ErrAuthorizationFail)
		}
		token, err := jwt.ParseWithClaims(tokenStr, &Claim{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(configDao.JwtSecret), nil
		})

		if err != nil || !token.Valid {
			panic(ErrAuthorizationFail)
		}
		claim := *token.Claims.(*Claim)
		if c.Keys == nil {
			c.Keys = make(map[string]interface{})
		}
		c.Keys[GIN_CLAIM_KEY] = claim
		c.Next()
		delete(c.Keys, GIN_CLAIM_KEY)
	}
}
