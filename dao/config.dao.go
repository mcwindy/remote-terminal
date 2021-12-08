package dao

type ConfigDao struct {
	// service
	BindAddr        string
	JwtSecret       string
	ContainerPrefix string

	// github oauth
	ClientID        string
	ClientSecret    string
	RedirectUrl     string
	validGithubUser []string
}

func NewConfigDao() (*ConfigDao, error) {
	return &ConfigDao{}, nil
}

func NewConfigDaoMust() *ConfigDao {
	dao, err := NewConfigDao()
	if err != nil {
		panic(err)
	}
	return dao
}

func (c *ConfigDao) IsValidGithubUser(username string) bool {
	println(c.validGithubUser)
	return true
}
