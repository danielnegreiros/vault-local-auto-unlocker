package storage

import (
	"log/slog"
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

	err = boltDB.db.Close()
	if err != nil {
		slog.Error("error closing boltdb", "error", err)
		os.Exit(1)
	}

	err = os.Remove(boltDB.path)
	if err != nil {
		slog.Error("error cleaning boltdb", "error", err)
		os.Exit(1)
	}
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
