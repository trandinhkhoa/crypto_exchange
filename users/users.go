package users

type User struct {
	Id         string
	PrivateKey string
	Balance    map[string]float64
}

type Wallet struct {
	PrivateKey string
	PublicKey  string
	Address    string
}

type Users map[string]*User
