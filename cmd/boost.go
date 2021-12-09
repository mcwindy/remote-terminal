package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ChenKS12138/remote-terminal/controller"
	"github.com/ChenKS12138/remote-terminal/dao"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"gopkg.in/yaml.v2"
)

type YamlConfig struct {
	Version int `yaml:"version"`
	Jwt     struct {
		Secret string `yaml:"secret"`
		Expire string `yaml:"expire"`
	} `yaml:"jwt"`
	Container struct {
		Prefix string `yaml:"prefix"`
	} `yaml:"container"`
	Oauth struct {
		Github struct {
			ClientID      string   `yaml:"clientID"`
			ClientSecret  string   `yaml:"clientSecret"`
			RedirectUrl   string   `yaml:"redirectUrl"`
			ValidLoginIDs []string `yaml:"validLoginIDs"`
		} `yaml:"github"`
	} `yaml:"oauth"`
}

func loadConfig() {
	configUrl := flag.String("config", "", "load online configuration")
	proxyUrl := flag.String("proxy", "", "net proxy")
	bindAddr := flag.String("bind", "", "bind addr")
	flag.Parse()
	if len(*configUrl) == 0 || len(*bindAddr) == 0 {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	log.Println("Ready To Fetch Config!")
	resp, err := resty.New().SetProxy(*proxyUrl).R().Get(*configUrl)

	if err != nil {
		panic(err)
	}

	yamlConfig := &YamlConfig{}

	if err = yaml.Unmarshal(resp.Body(), yamlConfig); err != nil {
		panic(err)
	}

	expire, err := time.ParseDuration(yamlConfig.Jwt.Expire)

	if err != nil {
		panic(err)
	}

	dao.Config.ClientID = yamlConfig.Oauth.Github.ClientID
	dao.Config.ClientSecret = yamlConfig.Oauth.Github.ClientSecret
	dao.Config.RedirectUrl = yamlConfig.Oauth.Github.RedirectUrl
	dao.Config.JwtSecret = yamlConfig.Jwt.Secret
	dao.Config.DiaProxy = *proxyUrl
	dao.Config.JwtExpire = expire
	dao.Config.BindAddr = *bindAddr
	dao.Config.ContainerPrefix = yamlConfig.Container.Prefix
	dao.Config.ValidGithubUser = make(map[string]interface{})
	for _, loginId := range yamlConfig.Oauth.Github.ValidLoginIDs {
		dao.Config.ValidGithubUser[loginId] = nil
	}
}

func Boost() {
	loadConfig()
	configDao := dao.NewConfigDaoMust()

	index := controller.NewIndexController()
	container := controller.NewContainerController()
	oauth := controller.NewOauthController()

	r := gin.New()
	r.LoadHTMLGlob("template/*.html")

	index.Group(r.Group("/"))
	container.Group(r.Group("/container"))
	oauth.Group(r.Group("/oauth"))

	srv := &http.Server{
		Addr:    configDao.BindAddr,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown: %s\n", err)
	}
	log.Println("Server Exit")
}
