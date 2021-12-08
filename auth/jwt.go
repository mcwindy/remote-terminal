package auth

import (
	"fmt"
	"strconv"

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
}

func GetClaim(c *gin.Context) *Claim {
	contextClaim := c.Keys[GIN_CLAIM_KEY].(jwt.MapClaims)
	permissionMask, _ := strconv.ParseUint(contextClaim["PermissionMask"].(string), 10, 64)
	claim := &Claim{
		GithubLogin:    contextClaim["GithubLogin"].(string),
		PermissionMask: permissionMask,
	}
	return claim
}

func SignClaim(claim *Claim) (string, error) {
	configDao := dao.NewConfigDaoMust()
	c := jwt.NewWithClaims(&jwt.SigningMethodHMAC{}, jwt.MapClaims{
		"GithubLogin":    claim.GithubLogin,
		"PermissionMask": fmt.Sprintf("%d", claim.PermissionMask),
	})
	return c.SignedString(configDao.JwtSecret)
}
