package dao

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type GithubDao struct {
	client resty.Client
}

func NewGithubDaoFromCode(code string) (*GithubDao, error) {
	configDao := NewConfigDaoMust()
	client := resty.New()
	resp, err := client.R().SetQueryParams(map[string]string{
		"client_id":     configDao.ClientID,
		"client_secret": configDao.ClientSecret,
		"code":          code,
	}).SetHeader("Accept", "application/json").
		Get("https://github.com/login/oauth/access_token")
	if err != nil {
		return nil, err
	}
	data := make(map[string]interface{})
	json.Unmarshal(resp.Body(), data)
	accessToken, accessTokenOk := data["access_token"].(string)
	tokenType, tokenTypeOk := data["token_type"].(string)
	if !accessTokenOk || !tokenTypeOk {
		return nil, errors.New("invalid code")
	}

	return &GithubDao{
		client: *resty.New().SetAuthScheme(tokenType).SetAuthToken(accessToken).SetBaseURL("https://api.github.com").OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
			if r.StatusCode() != http.StatusOK {
				return errors.New("status not ok")
			}
			return nil
		}),
	}, nil
}

func (g *GithubDao) GetLoginID() (string, error) {
	resp, err := g.client.R().Get("/user")
	if err != nil {
		return "", err
	}
	data := make(map[string]interface{})
	json.Unmarshal(resp.Body(), data)
	login, loginOk := data["login"].(string)
	if !loginOk {
		return "", errors.New("login not found")
	}
	return login, nil
}
