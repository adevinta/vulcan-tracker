/*
Copyright 2022 Adevinta
*/

// Package main start a vulcan-tracker server.
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/adevinta/vulcan-tracker/pkg/api"
	"github.com/adevinta/vulcan-tracker/pkg/config"
	"github.com/adevinta/vulcan-tracker/pkg/secrets"
	"github.com/adevinta/vulcan-tracker/pkg/storage/postgresql"
	"github.com/adevinta/vulcan-tracker/pkg/tracking"
)

func main() {
	configFilePath := flag.String("c", "_resources/config/local.toml", "configuration file")

	flag.Parse()
	cfg, err := config.ParseConfig(*configFilePath)
	if err != nil {
		log.Fatalf("Error reading configuration: %v", err)
	}

	e := echo.New()

	e.Logger.SetLevel(config.ParseLogLvl(cfg.Log.Level))

	// Database connection.
	cfgPSQL := cfg.PSQL
	db, err := postgresql.NewDB(cfgPSQL, e.Logger)
	if err != nil {
		e.Logger.Fatal(err)
	}

	secretsManager, err := secrets.NewAWSSecretManager(cfg.AWS, e.Logger)
	if err != nil {
		e.Logger.Fatal(err)
	}

	ticketServer, err := tracking.New(db, secretsManager, e.Logger)
	if err != nil {
		e.Logger.Fatal(err)
	}

	// Builder for the ticket tracker clients.
	ticketTrackerBuilder := &tracking.TTBuilder{}

	a := api.New(ticketServer, ticketTrackerBuilder, db, api.Options{
		DefaultTeamProject: cfg.API.DefaultTeamProject,
		MaxSize:            cfg.API.MaxSize,
		DefaultSize:        cfg.API.DefaultSize,
	})

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		// Avoid logging healthcheck requests.
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/healthcheck")
		},
	}))

	e.Use(middleware.Recover())
	e.Pre(middleware.RemoveTrailingSlash())

	e.GET("/:team_id/tickets/:id", a.GetTicket)
	e.POST("/:team_id/tickets", a.CreateTicket)
	e.GET("/:team_id/tickets/findings/:finding_id", a.GetFindingTicket)

	e.GET("/healthcheck", a.Healthcheck)

	address := fmt.Sprintf(":%d", cfg.API.Port)
	e.Logger.Fatal(e.Start(address))
}
