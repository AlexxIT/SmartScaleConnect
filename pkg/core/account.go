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

type AccountWithAddWeights interface {
	AddWeights(weights []*Weight) error
	DeleteWeight(weight *Weight) error
	Equal(a, b *Weight) bool
}
