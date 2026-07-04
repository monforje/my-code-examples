// Package app
package app

import (
	"context"
	"notifications/internal/app/closer"
	"notifications/internal/config"
	"notifications/internal/http/handlers"
	mailersender "notifications/internal/repository/mailer"
	postgresrepo "notifications/internal/repository/postgres"
	rediscache "notifications/internal/repository/redis"
	service "notifications/internal/services"
	cireportservice "notifications/internal/services/ci_report"
	reportservice "notifications/internal/services/report"
	"notifications/internal/templates"
	reportsclient "notifications/pkg/http_clients/reports"
	"notifications/pkg/logger"
	"notifications/pkg/mailer"
	natsclient "notifications/pkg/nats"
	postgresclient "notifications/pkg/postgres"
	redisclient "notifications/pkg/redis"
)

type diContainer struct {
	logger *logger.Logger
	ctx    context.Context
	cfg    *config.Config

	pg    *postgresclient.Client
	nats  *natsclient.Client
	smtp  *mailer.Client
	redis *redisclient.Client

	store    *postgresrepo.Store
	renderer *templates.Renderer
	sender   *mailersender.Sender

	notiSvc *service.NotificationService

	ciReportCache    *rediscache.ReportCache
	ciReportService  *cireportservice.CIReportService
	ciReportHandlers *handlers.CiReportHandlers

	reportsClient *reportsclient.ReportsClient
	reportService *reportservice.ReportService
}

func newDIContainer(ctx context.Context, logger *logger.Logger, cfg *config.Config) *diContainer {
	return &diContainer{
		logger: logger,
		ctx:    ctx,
		cfg:    cfg,
	}
}

func (d *diContainer) Postgres() *postgresclient.Client {
	if d.pg == nil {
		pg := postgresclient.New(d.ctx, d.cfg.PG, d.logger)
		closer.Add("postgres", func(ctx context.Context) error { pg.Close(); return nil })
		d.pg = pg
	}
	return d.pg
}

func (d *diContainer) NATS() *natsclient.Client {
	if d.nats == nil {
		d.nats = natsclient.New(d.ctx, d.cfg.NATS, d.logger)
		closer.Add("nats", func(ctx context.Context) error { d.nats.Close(ctx); return nil })
	}
	return d.nats
}

func (d *diContainer) SMTP() *mailer.Client {
	if d.smtp == nil {
		client, err := mailer.New(d.ctx, &d.cfg.SMTP)
		if err != nil {
			panic(err)
		}
		d.smtp = client
	}
	return d.smtp
}

func (d *diContainer) Redis() *redisclient.Client {
	if d.redis == nil {
		client := redisclient.New(d.ctx, d.cfg.Redis, d.logger)
		closer.Add("redis", func(ctx context.Context) error { client.Close(); return nil })
		d.redis = client
	}
	return d.redis
}

func (d *diContainer) Store() *postgresrepo.Store {
	if d.store == nil {
		d.store = postgresrepo.NewStore(postgresrepo.New(d.Postgres().Pool))
	}
	return d.store
}

func (d *diContainer) ProcessedEvents() *postgresrepo.ProcessedEventsRepo {
	return d.Store().ProcessedEvents()
}

func (d *diContainer) Renderer() *templates.Renderer {
	if d.renderer == nil {
		renderer, err := templates.NewRenderer()
		if err != nil {
			panic(err)
		}
		d.renderer = renderer
	}
	return d.renderer
}

func (d *diContainer) Sender() *mailersender.Sender {
	if d.sender == nil {
		d.sender = mailersender.New(
			mailersender.NewMailerAdapter(d.SMTP()),
			d.Renderer(),
			d.logger.Logger,
		)
	}
	return d.sender
}

func (d *diContainer) NotificationService() *service.NotificationService {
	if d.notiSvc == nil {
		d.notiSvc = service.NewNotificationService(d.Sender())
	}
	return d.notiSvc
}

func (d *diContainer) CIReportCache() *rediscache.ReportCache {
	if d.ciReportCache == nil {
		d.ciReportCache = rediscache.NewReportCache(d.Redis().Client, d.cfg.Features.CIReportCacheTTL)
	}
	return d.ciReportCache
}

func (d *diContainer) CIReportService() *cireportservice.CIReportService {
	if d.ciReportService == nil {
		d.ciReportService = cireportservice.NewCIReportService(d.CIReportCache(), d.ReportService(), d.logger, d.cfg.Features)
	}
	return d.ciReportService
}

func (d *diContainer) ReportsClient() *reportsclient.ReportsClient {
	if d.reportsClient == nil {
		d.reportsClient = reportsclient.NewReportsClient(d.cfg.HTTPClient.ReportsClient)
	}
	return d.reportsClient
}

func (d *diContainer) ReportService() *reportservice.ReportService {
	if d.reportService == nil {
		d.reportService = reportservice.NewReportService(d.ReportsClient(), d.logger)
	}
	return d.reportService
}

func (d *diContainer) CIReportHandlers() *handlers.CiReportHandlers {
	if d.ciReportHandlers == nil {
		d.ciReportHandlers = handlers.NewCiReportHandlers(d.CIReportService())
	}
	return d.ciReportHandlers
}
