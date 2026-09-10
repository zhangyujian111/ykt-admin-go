package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"ykt.dev/admin/internal/auth"
	"ykt.dev/admin/internal/db"
	"ykt.dev/admin/internal/handler"
	"ykt.dev/admin/internal/repo"
	"ykt.dev/admin/internal/router"
)

type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Mode string `yaml:"mode"`
	} `yaml:"server"`
	MySQL struct {
		DSN          string `yaml:"dsn"`
		MaxOpenConns int    `yaml:"maxOpenConns"`
		MaxIdleConns int    `yaml:"maxIdleConns"`
		ConnMaxLife  int    `yaml:"connMaxLifeMin"`
	} `yaml:"mysql"`
	Redis struct {
		Addr string `yaml:"addr"`
		DB   int    `yaml:"db"`
	} `yaml:"redis"`
	JWT struct {
		Secret string `yaml:"secret"`
		Issuer string `yaml:"issuer"`
		TTLMin int    `yaml:"ttlMin"`
	} `yaml:"jwt"`
	Log struct {
		Level string `yaml:"level"`
	} `yaml:"log"`
}

func loadConfig() *Config {
	cfg := &Config{}
	data, err := os.ReadFile("configs/config.yaml")
	if err != nil {
		slog.Error("read config", "err", err)
		os.Exit(1)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		slog.Error("parse config", "err", err)
		os.Exit(1)
	}
	return cfg
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	cfg := loadConfig()

	store, err := db.NewStore(
		db.Config{DSN: cfg.MySQL.DSN, MaxOpenConns: cfg.MySQL.MaxOpenConns, MaxIdleConns: cfg.MySQL.MaxIdleConns, ConnMaxLife: time.Duration(cfg.MySQL.ConnMaxLife) * time.Minute},
	)
	if err != nil {
		slog.Error("connect mysql", "err", err)
		os.Exit(1)
	}
	defer store.Close()
	slog.Info("mysql connected (single database)")

	jwtMgr := auth.New(cfg.JWT.Secret, cfg.JWT.Issuer, time.Duration(cfg.JWT.TTLMin)*time.Minute)

	deps := router.Deps{
		Store:   store,
		JWT:     jwtMgr,
		Users:   repo.NewUserRepo(store.Admin),
		Tenants: repo.NewTenantRepo(store.Admin),
		Alerts:  repo.NewAlertRepo(store.Admin),
		Devices: repo.NewDeviceRepo(store.Admin),
		Metrics: repo.NewMetricsRepo(store.Admin),
		System:  &handler.SystemHandler{Store: store},
	}

	r := router.New(deps)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	slog.Info("ykt-admin-go listening", "addr", addr)
	if err := r.Run(addr); err != nil {
		slog.Error("server exit", "err", err)
		os.Exit(1)
	}
}
