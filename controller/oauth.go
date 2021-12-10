package controller

import (
	"fmt"
	"net/http"

	"github.com/ChenKS12138/remote-terminal/auth"
	"github.com/ChenKS12138/remote-terminal/auth/permission"
	dao "github.com/ChenKS12138/remote-terminal/dao"
	"github.com/gin-gonic/gin"
)

type OauthController struct{}

func NewOauthController() *OauthController {
	return &OauthController{}
}

const COOKIE_EXPIRE = 7 * 24 * 60 * 60

func redirect(c *gin.Context) {
	code := c.Query("code")
	errorTitle, errorDescription, errorUri := c.Query("error"), c.Query("error_description"), c.Query("error_uri")

	if len(errorTitle) != 0 || len(errorDescription) != 0 || len(errorUri) != 0 {
		c.HTML(http.StatusOK, "error.html", gin.H{
			"error":       errorTitle,
			"description": errorDescription + errorUri,
		})
		return
	}
	if len(code) == 0 {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":       "error",
			"description": "description" + code,
		})
		return
	}

	githubDao, err := dao.NewGithubDaoFromCode(code)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/oauth/redirect?error=%s&error_description=%s", "invalid code", code))
		return
	}
	configDao := dao.NewConfigDaoMust()
	loginID, err := githubDao.GetLoginID()
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/oauth/redirect?error=%s&error_description=%s", "GetLoginID Fail", err.Error()))
		return
	}
	if !configDao.IsValidGithubUser(loginID) {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/oauth/redirect?error=%s&error_description=%s%s", "Unauthorized", loginID, ",You're Not In Whitelist!Wait for 10mins"))
		return
	}

	claim := &auth.Claim{
		GithubLogin:    loginID,
		PermissionMask: permission.RunOwnContainer,
	}
	jwtToken, err := auth.SignClaim(claim)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/oauth/redirect?error=%s&error_description=%s", "Sign Failed", err.Error()))
		return
	}
	c.SetCookie(auth.COOKIE_NAME, jwtToken, COOKIE_EXPIRE, "/", "", false, false)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (o *OauthController) Group(g *gin.RouterGroup) {
	g.GET("/redirect", redirect)
}
