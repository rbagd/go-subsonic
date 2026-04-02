package config

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadFromViperDefaults(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBufferString(`
server:
  url: http://example.com
  username: u
  password: p
`)); err != nil {
		t.Fatalf("ReadConfig error = %v", err)
	}

	cfg, err := loadFromViper(v)
	if err != nil {
		t.Fatalf("loadFromViper error = %v", err)
	}
	if cfg.Player.BufferSize != 10 {
		t.Fatalf("Player.BufferSize = %d, want 10", cfg.Player.BufferSize)
	}
}

func TestLoadFromViperParsesPlayer(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBufferString(`
server:
  url: http://example.com
  username: u
  password: p
player:
  buffer_size: 42
`)); err != nil {
		t.Fatalf("ReadConfig error = %v", err)
	}

	cfg, err := loadFromViper(v)
	if err != nil {
		t.Fatalf("loadFromViper error = %v", err)
	}
	if cfg.Player.BufferSize != 42 {
		t.Fatalf("Player.BufferSize = %d, want 42", cfg.Player.BufferSize)
	}
}
