package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/alpineworks/ootel"
	"github.com/michaelpeterswa/wikipedia-spider/internal/config"
	"github.com/michaelpeterswa/wikipedia-spider/internal/db"
	"github.com/michaelpeterswa/wikipedia-spider/internal/logging"
	"github.com/michaelpeterswa/wikipedia-spider/internal/spider"
)

func main() {
	slogHandler := slog.NewJSONHandler(os.Stdout, nil)
	slog.SetDefault(slog.New(slogHandler))

	slog.Info("welcome to wikipedia-spider!")

	c, err := config.NewConfig()
	if err != nil {
		slog.Error("could not create config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slogLevel, err := logging.LogLevelToSlogLevel(c.String(config.LogLevel))
	if err != nil {
		slog.Error("could not parse log level", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.SetLogLoggerLevel(slogLevel)

	ctx := context.Background()

	ootelClient := ootel.NewOotelClient(
		ootel.WithMetricConfig(
			ootel.NewMetricConfig(
				c.Bool(config.MetricsEnabled),
				c.Int(config.MetricsPort),
			),
		),
		ootel.WithTraceConfig(
			ootel.NewTraceConfig(
				c.Bool(config.TracingEnabled),
				c.Float64(config.TracingSampleRate),
				c.String(config.TracingService),
				c.String(config.TracingVersion),
			),
		),
	)

	shutdown, err := ootelClient.Init(ctx)
	if err != nil {
		slog.Error("could not initialize ootel client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer func() {
		_ = shutdown(ctx)
	}()

	dbClient, err := db.NewDBClient(ctx, c.String(config.PostgresConnString))
	if err != nil {
		slog.Error("could not create db client", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbClient.Close()

	jobID, err := dbClient.CreateJob(ctx)
	if err != nil {
		slog.Error("could not create job", slog.String("error", err.Error()))
		os.Exit(1)
	}

	spiderConfig := spider.NewSpiderConfig(
		c.String(config.SpiderUserAgent),
		strings.Split(c.String(config.SpiderAllowedDomains), ","),
		c.String(config.SpiderURLBase),
		c.String(config.SpiderStartSlug),
		c.Int(config.SpiderMaxDepth),
	)

	wikipediaSpider := spider.NewSpider(ctx, dbClient, spiderConfig, jobID)

	err = wikipediaSpider.Run()
	if err != nil {
		slog.Error("could not create spider", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("wikipedia-spider has finished! goodbye!")
}
