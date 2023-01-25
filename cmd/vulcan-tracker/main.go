/*
Copyright 2022 Adevinta
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/adevinta/vulcan-tracker/pkg/api"
	"github.com/adevinta/vulcan-tracker/pkg/config"
	"github.com/adevinta/vulcan-tracker/pkg/events"
	"github.com/adevinta/vulcan-tracker/pkg/events/kafka"
	"github.com/adevinta/vulcan-tracker/pkg/log"
	"github.com/adevinta/vulcan-tracker/pkg/model"
	"github.com/adevinta/vulcan-tracker/pkg/storage"
	"github.com/adevinta/vulcan-tracker/pkg/tracking"
	"github.com/adevinta/vulcan-tracker/pkg/vulcan"
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

	if err := log.SetLevel(strings.ToLower(cfg.Log.Level)); err != nil {
		log.Fatalf("error setting log level: %v", err)
	}

	e := echo.New()

	e.Logger.SetLevel(config.ParseLogLvl(cfg.Log.Level))

	// TODO: Decide which is the type of storage.
	storage, err := storage.New(cfg.Servers, cfg.Teams)
	if err != nil {
		e.Logger.Fatalf("Error initializing storage: %w", err)
	}

	serversConf, err := storage.ServersConf()
	if err != nil {
		e.Logger.Fatal(err)
	}

	// Generate a client for each type of tracker.
	trackerServers, err := tracking.GenerateServerClients(serversConf, e.Logger)
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

	e.GET("/:team_id/tickets/:id", a.GetTicket)
	e.POST("/:team_id/tickets", a.CreateTicket)
	e.POST("/:team_id/tickets/:id/fix", a.FixTicket)
	e.POST("/:team_id/tickets/:id/wontfix", a.WontFixTicket)

	errs := make(chan error, 1)

	go func() {
		if err := runEventReader(context.Background(), cfg.Kafka, storage, trackerServers); err != nil {
			errs <- fmt.Errorf("vulcan-tracker: %v", err)
		}
		// close(errs)
	}()
	if err := <-errs; err != nil {
		// handle error
		log.Error.Fatal(err)
	}

	address := fmt.Sprintf(":%d", cfg.API.Port)
	e.Logger.Fatal(e.Start(address))

}

// runEventReader is invoked by main and does the actual work of reading events.
func runEventReader(ctx context.Context, kCfg config.KafkaConfig, s storage.Storage, ts map[string]tracking.TicketTracker) error {

	kcfg := map[string]any{
		"bootstrap.servers":    kCfg.KafkaBroker,
		"group.id":             kCfg.KafkaGroupID,
		"auto.offset.reset":    "earliest",
		"enable.partition.eof": true,
	}

	if kCfg.KafkaUsername != "" && kCfg.KafkaPassword != "" {
		kcfg["security.protocol"] = "sasl_ssl"
		kcfg["sasl.mechanisms"] = "SCRAM-SHA-256"
		kcfg["sasl.username"] = kCfg.KafkaUsername
		kcfg["sasl.password"] = kCfg.KafkaPassword
	}

	proc, err := kafka.NewProcessor(kcfg)
	if err != nil {
		return fmt.Errorf("error creating kafka processor: %w", err)
	}
	defer proc.Close()

	vcli := vulcan.NewClient(proc)
	for {
		log.Info.Println("vulcan-tracker: processing findings")

		select {
		case <-ctx.Done():
			log.Info.Println("vulcan-tracker: context is done")
			return nil
		default:
		}

		if err := vcli.ProcessFindings(ctx, findingHandler(s, ts)); err != nil {
			err = fmt.Errorf("error processing findings: %w", err)
			if kCfg.RetryDuration == 0 {
				return err
			}
			log.Error.Printf("vulcan-tracker: %v", err)
		}

		log.Info.Printf("vulcan-tracker: retrying in %v", kCfg.RetryDuration)
		time.Sleep(kCfg.RetryDuration)
	}

}

// findingHandler processes finding events coming from a stream.
func findingHandler(s storage.Storage, ts map[string]tracking.TicketTracker) vulcan.FindingHandler {
	return func(payload events.FindingNotification, isNil bool) error {
		log.Debug.Printf("vulcan-tracker: payload=%#v isNil=%v", payload, isNil)
		for _, teamId := range payload.Target.Teams {
			pConfig, err := s.ProjectConfig(teamId)
			if err != nil {
				// The team doesn't have a configuration in vulcan-tracker. We'll skip its findings.
				continue
			}
			if !pConfig.AutoCreate {
				// This team doesn't have enable the auto create configuration.
				continue
			}
			err = upsertTicket(pConfig, ts[pConfig.ServerName], payload)
			if err != nil {
				return fmt.Errorf("could not upsert finding: %w", err)
			}
		}
		return nil
	}
}

func upsertTicket(pConfig *model.ProjectConfig, tt tracking.TicketTracker, payload events.FindingNotification) error {
	// Check if there is a ticket for this finding.
	ticket, err := tt.FindTicketByFindingID(pConfig.Project, pConfig.VulnerabilityIssueType, payload.ID)
	if err != nil {
		return fmt.Errorf("error searching the finding %s: %w", payload.ID, err)
	}

	// Create the ticket.
	if ticket == nil {
		// TODO: Define a good summary and description.
		ticket = &model.Ticket{
			Project:     pConfig.Project,
			TicketType:  pConfig.VulnerabilityIssueType,
			Summary:     payload.Issue.Summary,
			Description: fmt.Sprintf("FindingID: %s", payload.ID),
		}
		ticket, err := tt.CreateTicket(ticket)
		if err != nil {
			return fmt.Errorf("could not create a ticket: %w", err)
		}
		log.Debug.Printf("vulcan-tracker: created ticket %s ", ticket.Key)

		return nil

	} else {
		// Update the ticket.
		switch status := payload.Status; status {

		case "FIXED":
			_, err := tt.FixTicket(ticket.Key, pConfig.FixedWorkflow)
			if err != nil {
				return err
			}
		case "FALSE POSITIVE":
			_, err := tt.WontFixTicket(ticket.Key, pConfig.WontFixWorkflow, "Won't Do")
			if err != nil {
				return err
			}
		}
	}

	return nil
}
