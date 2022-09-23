/*
Copyright 2022 Adevinta
*/

package main

import (
	"flag"
	"log"
	"strings"

	"github.com/adevinta/vulcan-jira-api/pkg/api"
	mddleware "github.com/adevinta/vulcan-jira-api/pkg/api/middleware"
	"github.com/adevinta/vulcan-jira-api/pkg/issues/jira"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	configFilePath := flag.String("c", "_resources/config/local.toml", "configuration file")

	flag.Parse()

	config, err := parseConfig(*configFilePath)
	if err != nil {
		log.Fatalf("Error reading configuration: %v", err)
	}

	e := echo.New()

	e.Logger.SetLevel(parseLogLvl(config.Log.Level))

	is, err := jira.NewIS(config.Jira, e.Logger)
	if err != nil {
		e.Logger.Fatal(err)
	}

	a := api.New(is, api.Options{
		MaxSize:     config.API.MaxSize,
		DefaultSize: config.API.DefaultSize,
	})

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		// Avoid logging healthcheck requests
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/healthcheck")
		},
	}))

	e.Use(middleware.Recover())
	e.Use(mddleware.Authorization(mddleware.NewTagAuthorizer(is, e.Logger)))
	e.Pre(middleware.RemoveTrailingSlash())

	// e.GET("/issues", a.ListIssues)
	e.GET("/issues/:id", a.GetIssue)
	e.POST("/issues", a.CreateIssue)
	e.POST("/issues/:id/fixed", a.FixIssue)

	e.Logger.Fatal(e.Start(":8084"))

}
