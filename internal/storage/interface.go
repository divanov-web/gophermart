package storage

type Storage interface {
	Ping() error
}
