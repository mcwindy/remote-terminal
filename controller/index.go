package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type IndexController struct{}

func NewIndexController() *IndexController {
	return &IndexController{}
}

func (i *IndexController) index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "remote-terminal",
	})
}

func (i *IndexController) Group(g *gin.RouterGroup) {
	// configDao := dao.NewConfigDaoMust()
	// g.Use(middleware.Recover(func(c *gin.Context, i interface{}) {
	// 	c.Redirect(http.StatusFound, fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s", configDao.ClientID, configDao.RedirectUrl))
	// }), auth.RequireAuth(permission.None))
	g.GET("/", i.index)
}
