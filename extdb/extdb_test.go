package extdb

import (
	"testing"
	"time"
)

func TestLoadConfigUsesEnvironmentDefaults(t *testing.T) {
	t.Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/mbr?sslmode=disable")
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "11")
	t.Setenv("DATABASE_MAX_IDLE_CONNS", "7")
	t.Setenv("DATABASE_CONN_MAX_LIFETIME", "10m")
	t.Setenv("DATABASE_CONN_MAX_IDLE_TIME", "3m")

	cfg := LoadConfig()
	if cfg.DSN != "postgres://user:pass@localhost:5432/mbr?sslmode=disable" {
		t.Fatalf("expected dsn from env, got %q", cfg.DSN)
	}
	if cfg.MaxOpenConns != 11 {
		t.Fatalf("expected max open conns 11, got %d", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 7 {
		t.Fatalf("expected max idle conns 7, got %d", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != 10*time.Minute {
		t.Fatalf("expected conn max lifetime 10m, got %s", cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime != 3*time.Minute {
		t.Fatalf("expected conn max idle time 3m, got %s", cfg.ConnMaxIdleTime)
	}
}

func TestOpenRejectsNonPostgresDSN(t *testing.T) {
	_, err := Open(Config{DSN: "sqlite:///tmp/test.db"})
	if err == nil {
		t.Fatalf("expected invalid dsn error")
	}
}
