package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/appconfig"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

const (
	RedisSlug appconfig.LayerSlug = "redis"
	DBSlug    appconfig.LayerSlug = "db"
)

type AppSettings struct {
	Redis RedisSettings
	DB    DBSettings
}

type RedisSettings struct {
	Host string `glazed.parameter:"host"`
	Port int    `glazed.parameter:"port"`
}

type DBSettings struct {
	DSN string `glazed.parameter:"dsn"`
}

func mustLayer(layer layers.ParameterLayer, err error) layers.ParameterLayer {
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create layer: %v\n", err)
		os.Exit(1)
	}
	return layer
}

func main() {
	redisLayer := mustLayer(layers.NewParameterLayer(
		string(RedisSlug),
		"Redis",
		layers.WithPrefix("redis-"),
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"host",
				parameters.ParameterTypeString,
				parameters.WithDefault("127.0.0.1"),
			),
			parameters.NewParameterDefinition(
				"port",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(6379),
			),
		),
	))

	dbLayer := mustLayer(layers.NewParameterLayer(
		string(DBSlug),
		"Database",
		layers.WithPrefix("db-"),
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"dsn",
				parameters.ParameterTypeString,
				parameters.WithDefault("sqlite://./app.db"),
			),
		),
	))

	parser, err := appconfig.NewParser[AppSettings](
		// Programmatic values work well for libraries, tests, and integration code.
		appconfig.WithValuesForLayers(map[string]map[string]interface{}{
			string(RedisSlug): {"host": "cache.local", "port": 6380},
			string(DBSlug):    {"dsn": "postgres://localhost:5432/app"},
		}),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create parser: %v\n", err)
		os.Exit(1)
	}

	if err := parser.Register(RedisSlug, redisLayer, func(t *AppSettings) any { return &t.Redis }); err != nil {
		fmt.Fprintf(os.Stderr, "failed to register redis layer: %v\n", err)
		os.Exit(1)
	}
	if err := parser.Register(DBSlug, dbLayer, func(t *AppSettings) any { return &t.DB }); err != nil {
		fmt.Fprintf(os.Stderr, "failed to register db layer: %v\n", err)
		os.Exit(1)
	}

	cfg, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Redis: host=%s port=%d\n", cfg.Redis.Host, cfg.Redis.Port)
	fmt.Printf("DB: dsn=%s\n", cfg.DB.DSN)
}
