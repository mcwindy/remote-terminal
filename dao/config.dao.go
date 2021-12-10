package dao

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"gopkg.in/yaml.v2"
)

type ConfigDao struct {
	// service
	BindAddr        string
	JwtSecret       string
	JwtExpire       time.Duration
	ContainerPrefix string
	DiaProxy        string

	// github oauth
	ClientID        string
	ClientSecret    string
	RedirectUrl     string
	ValidGithubUser map[string]interface{}
}

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

var config ConfigDao
var configMutex sync.Mutex
var configExpire time.Time

var configUrl string

func pullConfig() (err error) {
	defer func() {
		r := recover()
		if r != nil {
			switch r.(type) {
			case error:
				err = r.(error)
			case string:
				err = errors.New(r.(string))
			default:
				err = errors.New("unexpected error")
			}
		}
	}()
	if configExpire.After(time.Now()) {
		return nil
	}
	configMutex.Lock()
	defer func() {
		configMutex.Unlock()
	}()
	log.Println("pulling config...")
	resp, err := resty.New().SetProxy(config.DiaProxy).R().Get(configUrl)

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

	config.ClientID = yamlConfig.Oauth.Github.ClientID
	config.ClientSecret = yamlConfig.Oauth.Github.ClientSecret
	config.RedirectUrl = yamlConfig.Oauth.Github.RedirectUrl
	config.JwtSecret = yamlConfig.Jwt.Secret
	config.JwtExpire = expire
	config.ContainerPrefix = yamlConfig.Container.Prefix
	config.ValidGithubUser = make(map[string]interface{})
	for _, loginId := range yamlConfig.Oauth.Github.ValidLoginIDs {
		config.ValidGithubUser[loginId] = nil
	}
	configExpire = time.Now().Add(time.Minute * 10)
	log.Println("pull config done!")
	return nil
}

func InitConfig(onlineConfigUrl string, proxyUrl string, bindAdd string) {
	config.DiaProxy = proxyUrl
	config.BindAddr = bindAdd
	configUrl = onlineConfigUrl
	configExpire = time.Now().Add(time.Second * (-1))
	pullConfig()
}

func NewConfigDao() (*ConfigDao, error) {
	if err := pullConfig(); err != nil {
		return nil, err
	}
	return &config, nil
}

func NewConfigDaoMust() *ConfigDao {
	dao, err := NewConfigDao()
	if err != nil {
		panic(err)
	}
	return dao
}

func (c *ConfigDao) IsValidGithubUser(username string) (ok bool) {
	_, ok = c.ValidGithubUser[username]
	return ok
}
