package controller

import (
	"net/http"

	"github.com/ChenKS12138/remote-terminal/auth"
	"github.com/ChenKS12138/remote-terminal/dao"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ContainerController struct{}

func NewContainerController() *ContainerController {
	return &ContainerController{}
}

func (cc *ContainerController) connect(c *gin.Context) {
	claim := auth.GetClaim(c)

	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	containerDao, err := dao.NewContainerDao()
	if err != nil {
		panic(err)
	}
	container, err := containerDao.FindByID(claim.GithubLogin)
	if err != nil {
		panic(err)
	}
	if container == nil {
		container, err = containerDao.CreateByID(claim.GithubLogin, ws.UnderlyingConn())
		if err != nil {
			panic(err)
		}
	}
	if err := containerDao.AttachAndWaitByID(container, ws.UnderlyingConn()); err != nil {
		panic(err)
	}

	// for {
	// 	mt, message, err := ws.ReadMessage()
	// 	if err != nil {
	// 		break
	// 	}
	// 	err = ws.WriteMessage(mt, message)
	// 	if err != nil {
	// 		break
	// 	}
	// }
}

func (c *ContainerController) Group(g *gin.RouterGroup) {
	// g.Use(middleware.Recover(func(c *gin.Context, i interface{}) {
	// 	c.JSON(http.StatusUnauthorized, gin.H{
	// 		"code":    -1,
	// 		"message": "unauthorized",
	// 	})
	// }), auth.RequireAuth(permission.RunContainer))
	g.GET("/connect", c.connect)
}
