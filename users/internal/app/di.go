// Package app
package app

import (
	"context"
	"users/internal/app/closer"
	"users/internal/config"
	"users/internal/http/handlers"
	"users/internal/http/middleware"
	postgresrepo "users/internal/repository/postgres"
	"users/internal/repository/security"
	"users/internal/repository/storage"
	service "users/internal/services"
	gitauthservice "users/internal/services/git_auth"
	gitauthclient "users/pkg/http_clients/git_auth"
	"users/pkg/logger"
	natsclient "users/pkg/nats"
	postgresclient "users/pkg/postgres"
)

type diContainer struct {
	logger *logger.Logger
	ctx    context.Context
	cfg    *config.Config

	pg   *postgresclient.Client
	nats *natsclient.Client

	store        *postgresrepo.Store
	tokenManager *security.Manager
	avatarStore  *storage.LocalAvatarStorage
	gitAuthHTTP  *gitauthclient.GitAuthClient

	usersSvc      *service.UsersService
	gitAuthSvc    *gitauthservice.GitAuthService
	usersHandlers *handlers.UsersHandlers
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

func (d *diContainer) Store() *postgresrepo.Store {
	if d.store == nil {
		d.store = postgresrepo.NewStore(postgresrepo.New(d.Postgres().Pool))
	}
	return d.store
}

func (d *diContainer) TokenManager() *security.Manager {
	if d.tokenManager == nil {
		d.tokenManager = security.NewManager(d.cfg.JWT.Secret, d.cfg.Features)
	}
	return d.tokenManager
}

func (d *diContainer) AvatarStorage() *storage.LocalAvatarStorage {
	if d.avatarStore == nil {
		d.avatarStore = storage.NewLocalAvatarStorage(d.cfg.Storage.AvatarDir, d.cfg.Storage.AvatarPublic)
	}
	return d.avatarStore
}

func (d *diContainer) ProcessedEvents() *postgresrepo.ProcessedEventsRepo {
	return d.Store().ProcessedEvents()
}

func (d *diContainer) GitAuthClient() *gitauthclient.GitAuthClient {
	if d.gitAuthHTTP == nil {
		d.gitAuthHTTP = gitauthclient.NewGitAuthClient(d.cfg.HttpClient.GitAuthClient)
	}
	return d.gitAuthHTTP
}

func (d *diContainer) GitAuthService() *gitauthservice.GitAuthService {
	if d.gitAuthSvc == nil {
		d.gitAuthSvc = gitauthservice.NewGitAuthService(
			d.logger,
			d.GitAuthClient(),
			d.Store().GitUsers(),
			d.Store().UserProfiles(),
		)
	}
	return d.gitAuthSvc
}

func (d *diContainer) AuthService() *service.UsersService {
	if d.usersSvc == nil {
		store := d.Store()
		d.usersSvc = service.NewUsersService(
			d.logger,
			store.UserProfiles(),
			d.AvatarStorage(),
			d.TokenManager(),
			d.GitAuthService(),
		)
	}
	return d.usersSvc
}

func (d *diContainer) UsersHandlers() *handlers.UsersHandlers {
	if d.usersHandlers == nil {
		d.usersHandlers = handlers.NewUsersHandlers(d.AuthService())
	}
	return d.usersHandlers
}

func (d *diContainer) BearerAuth() middleware.TokenValidator {
	return d.TokenManager()
}
