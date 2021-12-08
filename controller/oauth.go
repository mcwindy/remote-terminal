package controller

import (
	"net/http"

	"github.com/ChenKS12138/remote-terminal/auth"
	"github.com/ChenKS12138/remote-terminal/auth/permission"
	dao "github.com/ChenKS12138/remote-terminal/dao"
	"github.com/gin-gonic/gin"
)

type OauthController struct{}

const COOKIE_EXPIRE = 7 * 24 * 60 * 60

func token(c *gin.Context) {
	code := c.Query("code")
	errorTitle, errorDescription, errorUri := c.Query("error"), c.Query("error_description"), c.Query("error_uri")
	if len(code) == 0 {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":       "error",
			"description": "description",
		})
	}
	if len(errorTitle) != 0 || len(errorDescription) != 0 || len(errorUri) != 0 {
		c.HTML(http.StatusOK, "error.html", gin.H{
			"error":       errorTitle,
			"description": errorDescription + errorUri,
		})
	}

	githubDao, err := dao.NewGithubDaoFromCode(code)
	if err != nil {
		panic(err)
	}
	configDao := dao.NewConfigDaoMust()
	loginID, err := githubDao.GetLoginID()
	if err != nil {
		panic(err)
	}
	// validate fail
	if err != nil || !configDao.IsValidGithubUser(loginID) {
		// TODO clean cookie and render error
		// c.Redirect(http.StatusUnauthorized, fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s", configDao.ClientID, configDao.RedirectUrl))
	}

	claim := &auth.Claim{
		GithubLogin:    loginID,
		PermissionMask: permission.RunOwnContainer,
	}
	jwtToken, err := auth.SignClaim(claim)
	if err != nil {
		panic(err)
	}
	c.SetCookie(auth.COOKIE_NAME, jwtToken, COOKIE_EXPIRE, "/", "", false, false)
	c.Redirect(http.StatusOK, "/")
}

func (o *OauthController) Group(g *gin.RouterGroup) {
	g.GET("/token", token)
}
