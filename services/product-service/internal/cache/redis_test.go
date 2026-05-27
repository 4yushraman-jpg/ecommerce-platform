package cache

import (
	"testing"

	"github.com/alicebob/miniredis/v2"

	"github.com/4yushraman-jpg/product-service/internal/config"
)

func TestNewRedis(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	cfg := &config.Config{RedisAddr: mr.Addr(), RedisDB: 0}
	client, err := NewRedis(cfg)
	if err != nil {
		t.Fatalf("new redis: %v", err)
	}
	defer client.Close()
}
