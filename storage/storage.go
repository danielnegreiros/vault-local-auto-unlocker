package storage

type Storage interface {
	RetrieveKey(table string, key string) (string, error)
	InsertKeyValue(table string, key string, data string) error
}
