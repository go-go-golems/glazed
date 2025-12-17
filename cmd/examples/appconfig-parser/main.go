package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/appconfig"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
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

func mustSection(section *schema.SectionImpl, err error) schema.Section {
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create section: %v\n", err)
		os.Exit(1)
	}
	return section
}

func main() {
	redisLayer := mustSection(schema.NewSection(
		string(RedisSlug),
		"Redis",
		schema.WithPrefix("redis-"),
		schema.WithFields(
			fields.New("host", fields.TypeString, fields.WithDefault("127.0.0.1")),
			fields.New("port", fields.TypeInteger, fields.WithDefault(6379)),
		),
	))

	dbLayer := mustSection(schema.NewSection(
		string(DBSlug),
		"Database",
		schema.WithPrefix("db-"),
		schema.WithFields(
			fields.New("dsn", fields.TypeString, fields.WithDefault("sqlite://./app.db")),
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
