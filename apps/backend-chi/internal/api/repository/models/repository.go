package models

type Repository struct {
	User  UserRepository
	Link  LinkRepository
	State StateStoreRepository
}

func NewRepository(userRepo UserRepository, linkRepo LinkRepository, stateStore StateStoreRepository) Repository {
	return Repository{
		User:  userRepo,
		Link:  linkRepo,
		State: stateStore,
	}
}
