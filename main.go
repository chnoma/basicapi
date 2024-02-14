package main

// This API is a proof of concept only. It represents a -very- simplified vendor platform.
// It does not actually do anything, nor does it have any security implemented.

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

type ErrorResponse struct {
	Message string `json:"error"`
}

type Server struct {
	Pg *pgxpool.Pool
}

type config struct {
	Postgres struct {
		Url string `yaml:"url"`
	} `yaml:"postgres"`
}

func main() {
	f, err := os.Open("./config.yml")
	if err != nil {
		panic(err)
	}

	var cfg config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		panic(err)
	}

	f.Close()

	pg, err := pgxpool.New(context.TODO(), cfg.Postgres.Url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer pg.Close()

	server := Server{pg}

	e := echo.New()
	server.setup_order_data_routes(e)
	server.setup_product_data_routes(e)
	server.setup_order_hypermedia_routes(e)
	server.setup_product_hypermedia_routes(e)

	e.Logger.Fatal(e.Start(":8000"))
}
