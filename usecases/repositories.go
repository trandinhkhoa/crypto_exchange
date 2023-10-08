package usecases

import "github.com/trandinhkhoa/crypto-exchange/entities"

type OrdersRepository interface {
	Create(entities.Order)
	// Read()
	Update(entities.Order)
	Delete(entities.Order)
}

type UsersRepository interface {
	Create(entities.User)
	// Read()
	Update(entities.User)
	// Delete()
}

// TODO: Last Trade Repository
