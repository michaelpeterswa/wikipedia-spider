package spider

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gocolly/colly"
	"github.com/google/uuid"
	"github.com/michaelpeterswa/wikipedia-spider/internal/db"
	"github.com/michaelpeterswa/wikipedia-spider/internal/rules"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

var (
	meter = otel.Meter("github.com/michaelpeterswa/wikipedia-spider/internal/spider") // meter is the named meter for the spider package

	sitesVisitedMetric otelmetric.Int64Counter // sitesVisitedMetric is the counter for the number of sites visited
	linksFoundMetric   otelmetric.Int64Counter // linksFoundMetric is the counter for the number of links found while visiting a site
	visitErrorMetric   otelmetric.Int64Counter // visitErrorMetric is the counter for the number of errors encountered while visiting a site
)

func init() {
	var err error

	sitesVisitedMetric, err = meter.Int64Counter("sites_visited")
	if err != nil {
		slog.Error("failed to create sitesVisited metric", slog.String("err", err.Error()))
		os.Exit(1)
	}

	linksFoundMetric, err = meter.Int64Counter("links_found")
	if err != nil {
		slog.Error("failed to create linksFound metric", slog.String("err", err.Error()))
		os.Exit(1)
	}

	visitErrorMetric, err = meter.Int64Counter("visit_error")
	if err != nil {
		slog.Error("failed to create pageAlreadyVisited metric", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

type SpiderConfig struct {
	UserAgent      string
	AllowedDomains []string
	URLBase        string
	StartSlug      string
	MaxDepth       int
}

func NewSpiderConfig(ua string, ad []string, ub string, s string, md int) *SpiderConfig {
	return &SpiderConfig{
		UserAgent:      ua,
		AllowedDomains: ad,
		URLBase:        ub,
		StartSlug:      s,
		MaxDepth:       md,
	}
}

type Spider struct {
	spider *colly.Collector
	config *SpiderConfig
}

func NewSpider(ctx context.Context, db *db.DBClient, sc *SpiderConfig, job_id *uuid.UUID) *Spider {
	spider := colly.NewCollector(
		colly.UserAgent(sc.UserAgent),
		colly.AllowedDomains(sc.AllowedDomains...),
		colly.MaxDepth(sc.MaxDepth),
	)

	spider.OnRequest(func(r *colly.Request) {
		slog.Info("visiting", slog.String("url", r.URL.String()))

		sitesVisitedMetric.Add(ctx, 1)
	})

	spider.OnHTML("span.mw-page-title-main", func(e *colly.HTMLElement) {
		slog.Debug("found title", slog.String("title", e.Text))
		e.Request.Ctx.Put("page-title", e.Text)

		// check if the link is already in the database
		link, err := db.GetLinkFromURL(ctx, e.Request.URL.String())
		if err != nil {
			slog.Error("could not get link from url", slog.String("url", e.Request.URL.String()), slog.String("error", err.Error()))
		}

		if link == nil {
			linkID := uuid.New()
			err := db.InsertLink(ctx, &linkID, job_id, e.Request.URL.String(), e.Text)
			if err != nil {
				slog.Error("could not insert link", slog.String("url", e.Request.URL.String()), slog.String("error", err.Error()))
			}
		}
	})

	spider.OnHTML("p", func(e *colly.HTMLElement) {
		anchors := e.ChildAttrs("a", "href")

		for _, anchor := range anchors {
			if rules.IsValidLink(anchor) {
				slog.Debug("found link", slog.String("link", anchor), slog.String("page-title", e.Request.Ctx.Get("page-title")))
				linksFoundMetric.Add(ctx, 1, otelmetric.WithAttributes(attribute.String("page-title", e.Request.Ctx.Get("page-title"))))

				fullLink := sc.URLBase + anchor

				link, err := db.GetLinkFromURL(ctx, fullLink)
				if err != nil {
					slog.Error("could not get link from url", slog.String("url", fullLink), slog.String("error", err.Error()))
				}

				if link == nil {
					linkID := uuid.New()
					err := db.InsertLink(ctx, &linkID, job_id, fullLink, e.Request.Ctx.Get("page-title"))
					if err != nil {
						slog.Error("could not insert link", slog.String("url", fullLink), slog.String("error", err.Error()))
					}
				}

				err = e.Request.Visit(anchor)
				if err != nil {
					if err == colly.ErrAlreadyVisited || err == colly.ErrMaxDepth {
						visitErrorMetric.Add(ctx, 1, otelErrorAttribute(err))
						slog.Debug("page already visited", slog.String("url", anchor))
						continue
					}
					visitErrorMetric.Add(ctx, 1, otelErrorAttribute(err))
					slog.Error("could not visit page", slog.String("error", err.Error()))
					continue
				}
			}
		}
	})

	err := spider.Limit(&colly.LimitRule{
		Delay: 10 * time.Second,
	})
	if err != nil {
		slog.Error("could not set rate limit", slog.String("error", err.Error()))
	}

	return &Spider{
		spider: spider,
		config: sc,
	}
}

func (s *Spider) Run() error {
	err := s.spider.Visit(s.config.URLBase + s.config.StartSlug)
	if err != nil {
		return fmt.Errorf("could not visit start slug: %w", err)
	}

	return nil
}

func otelErrorAttribute(err error) otelmetric.MeasurementOption {
	return otelmetric.WithAttributes(attribute.String("error", err.Error()))
}
