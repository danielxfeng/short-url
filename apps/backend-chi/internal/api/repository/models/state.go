package models

type StateStoreRepository interface {
	Add(state string, verifier string)
	GetAndDelete(state string) string
}
