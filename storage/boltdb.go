package storage

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/user"
	"strings"
	"vault-unlocker/conf"

	"go.etcd.io/bbolt"
)

type BoltBDStorage struct {
	path string
	db   *bbolt.DB
}

// UpdateKeys implements Storage.
func (b *BoltBDStorage) UpdateKeys() {
	panic("unimplemented")
}

func NewBoltDBStorage(boltConf *conf.BoltBD) (*BoltBDStorage, error) {

	parts := strings.Split(boltConf.Path, "/")
	path := strings.Join(parts[:len(parts)-1], "/")

	err := ensurePath(path)
	if err != nil {
		return nil, fmt.Errorf("initializing boldDB directory: [%w]", err)
	}

	db, err := bbolt.Open(boltConf.Path, 0666, nil)
	if err != nil {
		return nil, err
	}

	// Create bucket if not exists
	for _, bucketName := range bucketsName {
		err = db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
			return err
		})
		if err != nil {
			return nil, err
		}
	}

	bolt := BoltBDStorage{
		path: boltConf.Path,
		db:   db,
	}

	return &bolt, nil
}

var _ Storage = (*BoltBDStorage)(nil)

var bucketsName = []string{"users", "keys"}

func (b *BoltBDStorage) InsertKeyValue(table string, key string, data string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(table))
		return b.Put([]byte(key), []byte(data))
	})
}

// RetrieveKeys implements Storage.
func (b *BoltBDStorage) RetrieveKey(table string, key string) (string, error) {
	var data []byte

	err := b.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(table))
		data = b.Get([]byte(key))
		if data == nil {
			return errors.New("key not found")
		}
		return nil
	})
	return string(data), err
}

func ensurePath(path string) error {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		err := os.MkdirAll(path, 0750)
		if err != nil {
			u, _ := user.Current()
			dir, _ := os.Getwd()
			slog.Error("error", "user", u.Name, "home", u.HomeDir, "pwd", dir)
			return fmt.Errorf("failed to create directory: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking path: %w", err)
	}

	return nil
}
