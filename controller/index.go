package controller

import (
	"fmt"
	"net/http"

	"github.com/ChenKS12138/remote-terminal/auth"
	"github.com/ChenKS12138/remote-terminal/auth/permission"
	dao "github.com/ChenKS12138/remote-terminal/dao"
	"github.com/ChenKS12138/remote-terminal/middleware"
	"github.com/gin-gonic/gin"
)

type IndexController struct{}

func NewIndexController() *IndexController {
	return &IndexController{}
}

func (i *IndexController) index(c *gin.Context) {
	claim := auth.GetClaim(c)
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": fmt.Sprintf("%s@remote", claim.GithubLogin),
	})
}

const SCOPE = "read:user"

func (i *IndexController) Group(g *gin.RouterGroup) {
	configDao := dao.NewConfigDaoMust()
	g.Use(middleware.Recover(func(c *gin.Context, i interface{}) {
		c.SetCookie(auth.COOKIE_NAME, "", -1, "/", "", false, false)
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s", configDao.ClientID, configDao.RedirectUrl, SCOPE))
	}), auth.RequireAuth(permission.RunOwnContainer))
	g.GET("/", i.index)
}
