package storage

import (
	"os"
	"testing"
	"vault-unlocker/conf"

	"github.com/stretchr/testify/assert"
)

var boltDB *BoltBDStorage

func TestMain(m *testing.M) {
	var data = []byte(`
storage:
  type: boltdb
  boltdb:
    path: ../temp/bolt.db
`)

	cfg, err := conf.NewConfig(data)
	if err != nil {
		panic(err)
	}

	boltDB, err = NewBoltDBStorage(cfg.Storage.BoltDB)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	boltDB.db.Close()
	os.Remove(boltDB.path)
	os.Exit(code)
}

func TestBoltDB(t *testing.T) {
	err := boltDB.InsertKeyValue("keys", "0", "somerandomkey")
	assert.NoError(t, err)

	value, err := boltDB.RetrieveKey("keys", "0")
	assert.NoError(t, err)
	assert.Equal(t, "somerandomkey", value)

	_, err = boltDB.RetrieveKey("keys", "1")
	assert.ErrorContains(t, err, "not found")
}
