package auth

import (
	"time"

	"github.com/ChenKS12138/remote-terminal/dao"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

const COOKIE_NAME = "jwt_token"
const GIN_CLAIM_KEY = "claims"

type PermissionType = uint64

type Claim struct {
	GithubLogin    string
	PermissionMask PermissionType
	jwt.StandardClaims
}

func GetClaim(c *gin.Context) Claim {
	contextClaim := c.Keys[GIN_CLAIM_KEY].(Claim)
	return contextClaim
}

func SignClaim(claim *Claim) (string, error) {
	configDao := dao.NewConfigDaoMust()
	claim.ExpiresAt = time.Now().Add(time.Hour * 1).Unix()
	claim.IssuedAt = time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString([]byte(configDao.JwtSecret))
}
