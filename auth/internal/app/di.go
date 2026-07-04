package app

import (
	"auth/internal/app/closer"
	"auth/internal/config"
	"auth/internal/http/handlers"
	authclihandlers "auth/internal/http/handlers/auth_cli"
	"auth/internal/http/middleware"
	natsrepo "auth/internal/repository/nats"
	postgresrepo "auth/internal/repository/postgres"
	redisrepo "auth/internal/repository/redis"
	"auth/internal/repository/security"
	authservice "auth/internal/services/auth"
	kafkaclient "auth/pkg/kafka"
	"auth/pkg/logger"
	natsclient "auth/pkg/nats"
	postgresclient "auth/pkg/postgres"
	redisclient "auth/pkg/redis"
	"context"
)

type diContainer struct {
	logger *logger.Logger
	ctx    context.Context
	cfg    *config.Config

	pg    *postgresclient.Client
	redis *redisclient.Client
	kafka *kafkaclient.Client
	nats  *natsclient.Client

	store         *postgresrepo.Store
	rateLimiter   *redisrepo.RateLimiter
	kafkaProducer *kafkaclient.Client
	natsProducer  *natsrepo.Producer
	tokenManager  *security.Manager

	authSvc        *authservice.AuthService
	authHandlers   *handlers.AuthHandlers
	authCliHandlers *authclihandlers.AuthCliHandlers
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

func (d *diContainer) Redis() *redisclient.Client {
	if d.redis == nil {
		client := redisclient.New(d.ctx, d.cfg.Redis)
		closer.Add("redis", func(ctx context.Context) error { client.Close(); return nil })
		d.redis = client
	}
	return d.redis
}

func (d *diContainer) Kafka() *kafkaclient.Client {
	if d.kafka == nil {
		d.kafka = kafkaclient.New(d.ctx, d.cfg.Kafka, d.logger)
		closer.Add("kafka", func(ctx context.Context) error { d.kafka.Close(); return nil })
	}
	return d.kafka
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

func (d *diContainer) RateLimiter() *redisrepo.RateLimiter {
	if d.rateLimiter == nil {
		d.rateLimiter = redisrepo.NewRateLimiter(d.Redis().Client)
	}
	return d.rateLimiter
}

func (d *diContainer) Producer() *natsrepo.Producer {
	if d.natsProducer == nil {
		d.natsProducer = natsrepo.NewProducer(d.NATS().Conn())
	}
	return d.natsProducer
}

func (d *diContainer) TokenManager() *security.Manager {
	if d.tokenManager == nil {
		d.tokenManager = security.NewManager(d.cfg.JWT.Secret, d.cfg.Features)
	}
	return d.tokenManager
}

func (d *diContainer) TransactionFunc() authservice.TransactionFunc {
	store := d.Store()
	return func(ctx context.Context, fn func(authservice.Repositories) error) error {
		return store.WithTx(ctx, func(txStore *postgresrepo.Store) error {
			repos := authservice.Repositories{
				Identities:                txStore.Identities(),
				Credentials:               txStore.Credentials(),
				Sessions:                  txStore.Sessions(),
				VerificationCodes:         txStore.VerificationCodes(),
				PasswordResetTokens:       txStore.PasswordResetTokens(),
				PasswordChangeTokens:      txStore.PasswordChangeTokens(),
				EmailChangeRequests:       txStore.EmailChangeRequests(),
				AccountDeleteRequests:     txStore.AccountDeleteRequests(),
				AuthEvents:                txStore.AuthEvents(),
				DeviceAuthorizationCodes:  txStore.DeviceAuthorizationCodes(),
			}
			return fn(repos)
		})
	}
}

func (d *diContainer) AuthService() *authservice.AuthService {
	if d.authSvc == nil {
		store := d.Store()
		d.authSvc = authservice.NewAuthService(
			store.Identities(),
			store.Credentials(),
			store.Sessions(),
			store.VerificationCodes(),
			store.PasswordResetTokens(),
			store.PasswordChangeTokens(),
			store.EmailChangeRequests(),
			store.AccountDeleteRequests(),
			store.AuthEvents(),
			store.DeviceAuthorizationCodes(),
			d.Producer(),
			d.TokenManager(),
			d.TransactionFunc(),
			d.RateLimiter(),
			d.cfg.Features,
			d.cfg.VerificationURL,
		)
	}
	return d.authSvc
}

func (d *diContainer) AuthHandlers() *handlers.AuthHandlers {
	if d.authHandlers == nil {
		d.authHandlers = handlers.NewAuthHandlers(d.AuthService(), d.cfg.Features.RefreshSessionTTL)
	}
	return d.authHandlers
}

func (d *diContainer) AuthCliHandlers() *authclihandlers.AuthCliHandlers {
	if d.authCliHandlers == nil {
		d.authCliHandlers = authclihandlers.NewAuthCliHandlers(d.AuthService(), d.cfg.Features.RefreshSessionTTL)
	}
	return d.authCliHandlers
}

func (d *diContainer) BearerAuth() middleware.TokenValidator {
	return d.TokenManager()
}
