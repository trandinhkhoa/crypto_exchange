package usecases

import "github.com/trandinhkhoa/crypto-exchange/entities"

type OrdersRepository interface {
	Create(entities.Order)
	// Read()
	ReadAll(string) []entities.Order
	Update(entities.Order)
	Delete(entities.Order)
}

type UsersRepository interface {
	Create(entities.User)
	ReadAll() []entities.User
	Update(entities.User)
	// Delete()
}

// TODO: Last Trade Repository
type LastTradesRepository interface {
	Create(entities.Trade)
	ReadAll() []entities.Trade
}
