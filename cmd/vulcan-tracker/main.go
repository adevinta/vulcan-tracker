/*
Copyright 2022 Adevinta
*/

package main

import (
	"flag"
	"fmt"
	"github.com/adevinta/vulcan-tracker/pkg/storage/postgresql"
	"log"
	"strings"

	"github.com/adevinta/vulcan-tracker/pkg/api"
	"github.com/adevinta/vulcan-tracker/pkg/config"
	"github.com/adevinta/vulcan-tracker/pkg/storage"
	"github.com/adevinta/vulcan-tracker/pkg/tracking"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	// TODO: Decide which is the type of storage.
	ticketServerStorage, err := storage.New(cfg.Servers, cfg.Teams)
	if err != nil {
		e.Logger.Fatalf("Error initializing storage: %w", err)
	}

	serversConf, err := ticketServerStorage.ServersConf()
	if err != nil {
		e.Logger.Fatal(err)
	}

	// Generate a client for each type of tracker.
	trackerServers, err := tracking.GenerateServerClients(serversConf, e.Logger)
	if err != nil {
		e.Logger.Fatal(err)
	}

	cfgPSQL := cfg.PSQL
	cfgPSQLRead := cfg.PSQLRead
	// If we don't have defined the read replica we use the read write one.
	if cfgPSQLRead.Host == "" {
		cfgPSQLRead = cfgPSQL
	}
	db, err := postgresql.NewDB(cfgPSQL, cfgPSQLRead, e.Logger)
	if err != nil {
		e.Logger.Fatal(err)
	}

	a := api.New(trackerServers, ticketServerStorage, db, api.Options{
		MaxSize:     cfg.API.MaxSize,
		DefaultSize: cfg.API.DefaultSize,
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
	e.POST("/:team_id/tickets/:id/fix", a.FixTicket)
	e.POST("/:team_id/tickets/:id/wontfix", a.WontFixTicket)

	e.GET("/:team_id/tickets/findings/:finding_id", a.GetFindingTicket)
	address := fmt.Sprintf(":%d", cfg.API.Port)
	e.Logger.Fatal(e.Start(address))

}
