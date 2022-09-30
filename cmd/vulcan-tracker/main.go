/*
Copyright 2022 Adevinta
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/adevinta/vulcan-tracker/pkg/api"
	"github.com/adevinta/vulcan-tracker/pkg/config"
	"github.com/adevinta/vulcan-tracker/pkg/storage/toml"
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
	storage, err := toml.New(cfg.Servers, cfg.Teams)
	if err != nil {
		e.Logger.Fatal(err)
	}

	trackerServersConf, err := storage.ListTrackerServersConf()
	if err != nil {
		e.Logger.Fatal(err)
	}

	// Generate a client for each type of tracker.
	trackerServers, err := tracking.GenerateServerClients(trackerServersConf, e.Logger)
	if err != nil {
		e.Logger.Fatal(err)
	}

	a := api.New(trackerServers, storage, api.Options{
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

	e.GET("/tickets/:id", a.GetTicket)
	e.POST("/tickets", a.CreateTicket)
	e.POST("/tickets/:id/fixed", a.FixTicket)

	address := fmt.Sprintf(":%d", cfg.API.Port)
	e.Logger.Fatal(e.Start(address))

}
