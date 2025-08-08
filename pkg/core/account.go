package core

type Account interface {
	Login(username, password string) error
	GetAllWeights() ([]*Weight, error)
}

type AccountWithToken interface {
	Account
	LoginWithToken(token string) error
	Token() string
}

type AccountWithFilter interface {
	GetFilterWeights(name string) ([]*Weight, error)
}
