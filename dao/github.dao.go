package dao

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

type GithubDao struct {
	client resty.Client
}

func NewGithubDaoFromCode(code string) (*GithubDao, error) {
	configDao := NewConfigDaoMust()
	client := resty.New().SetProxy(configDao.DiaProxy)
	resp, err := client.R().
		SetBody(fmt.Sprintf(`{"client_id":"%s", "client_secret":"%s", "code": "%s"}`, configDao.ClientID, configDao.ClientSecret, code)).
		SetHeaders(map[string]string{
			"Accept":       "application/json",
			"Content-Type": "application/json",
		}).
		Post("https://github.com/login/oauth/access_token")

	if err != nil {
		return nil, err
	}
	data := &struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}{}
	err = json.Unmarshal(resp.Body(), data)
	if err != nil {
		return nil, err
	}

	return &GithubDao{
		client: *resty.New().SetProxy(configDao.DiaProxy).SetAuthScheme(data.TokenType).SetAuthToken(data.AccessToken).SetBaseURL("https://api.github.com").SetHeader("accept", "application/json"),
	}, nil
}

func (g *GithubDao) GetLoginID() (loginId string, err error) {
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
			loginId = ""
		}
	}()
	resp, err := g.client.R().Get("/user")
	if err != nil {
		return "", err
	}
	data := make(map[string]interface{})
	json.Unmarshal(resp.Body(), &data)
	login, loginOk := data["login"].(string)
	if !loginOk {
		return "", errors.New("login not found")
	}
	return login, nil
}
