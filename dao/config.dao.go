package dao

import "time"

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

var Config ConfigDao

func NewConfigDao() (*ConfigDao, error) {
	return &Config, nil
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
