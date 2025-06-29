package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"vault-unlocker/conf"
	"vault-unlocker/storage"
	vault_manager "vault-unlocker/vault"
)

const (
	confPath = "./examples/config.yaml"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	slog.SetDefault(logger)

	var content []byte
	var err error

	confPath := os.Getenv("CONF_PATH")
	if confPath != "" {
		content, err = conf.ReadFile(confPath)
		if err != nil {
			slog.Error("load config", "err", err)
		}

	}

	c, err := conf.NewConfig(content)
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	store, err := storage.NewBoltDBStorage(c.Storage.BoltDB)
	if err != nil {
		slog.Error("storage config", "err", err)
		os.Exit(1)
	}

	vClient, err := vault_manager.NewVaultClient(c.Unlocker)
	if err != nil {
		slog.Error("init vault client", "err", err)
		os.Exit(1)
	}

	vm, err := vault_manager.NewVaultManager(c.Unlocker, vClient, store)
	if err != nil {
		slog.Error("unlocker config", "err", err)
		os.Exit(1)
	}

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	endChan := make(chan os.Signal, 1)
	signal.Notify(endChan, syscall.SIGINT, syscall.SIGTERM)

	if err := vm.Unlock(ctx); err != nil {
		slog.Error("vault manager", "err", err)
	}

	wg := &sync.WaitGroup{}

	for {
		select {
		case <-ticker.C:
			go func() {
				wg.Add(1)
				defer wg.Done()
				opCtx, cancel := context.WithTimeout(ctx, 50*time.Second)
				defer cancel()
				if err := vm.Unlock(opCtx); err != nil {
					slog.Error("vault manager", "err", err)
				}
			}()
		case <-endChan:
			slog.Warn("received interruption signal")
			stop()
			wg.Wait()
			return
		case <-ctx.Done():
			slog.Info("shutting down")
			return
		}
	}
}
