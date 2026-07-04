package app

import (
	"context"

	"tasks/internal/app/closer"
	"tasks/internal/config"
	"tasks/internal/http/handlers"
	gittaskshandler "tasks/internal/http/handlers/git_tasks"
	reportshandler "tasks/internal/http/handlers/reports"
	postgresrepo "tasks/internal/repository/postgres"
	"tasks/internal/repository/security"
	taskservice "tasks/internal/services"
	gittasksservice "tasks/internal/services/git_tasks"
	reportsservice "tasks/internal/services/reports"
	gittasksclient "tasks/pkg/http_clients/git_tasks"
	usersserviceclient "tasks/pkg/http_clients/users_service"
	"tasks/pkg/logger"
	postgresclient "tasks/pkg/postgres"
)

type diContainer struct {
	logger *logger.Logger
	ctx    context.Context
	cfg    *config.Config

	pg *postgresclient.Client

	store        *postgresrepo.Store
	tokenManager *security.Manager

	tasksSvc        *taskservice.TasksService
	tagsSvc         *taskservice.TagsService
	languagesSvc    *taskservice.LanguagesService
	gitTasksSvc     *gittasksservice.GitTasksService
	reportsSvc      *reportsservice.ReportsService
	usersClient     *usersserviceclient.UsersClient
	gitTasksClient  *gittasksclient.GitAuthClient
	tasksHandlers   *handlers.TasksHandlers
	gitTasksHandler *gittaskshandler.GitTasksHandler
	reportsHandler  *reportshandler.ReportsHandler
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

func (d *diContainer) TasksService() *taskservice.TasksService {
	if d.tasksSvc == nil {
		d.tasksSvc = taskservice.NewTasksService(d.Store().Tasks())
	}
	return d.tasksSvc
}

func (d *diContainer) TagsService() *taskservice.TagsService {
	if d.tagsSvc == nil {
		d.tagsSvc = taskservice.NewTagsService(d.Store().Tasks())
	}
	return d.tagsSvc
}

func (d *diContainer) LanguagesService() *taskservice.LanguagesService {
	if d.languagesSvc == nil {
		d.languagesSvc = taskservice.NewLanguagesService(d.Store().Tasks())
	}
	return d.languagesSvc
}

func (d *diContainer) GitTasksService() *gittasksservice.GitTasksService {
	if d.gitTasksSvc == nil {
		d.gitTasksSvc = gittasksservice.NewGitTasksService(
			d.Store().Tasks(),
			d.Store().PulledTasks(),
			d.UsersClient(),
			d.GitTasksClient(),
		)
	}
	return d.gitTasksSvc
}

func (d *diContainer) ReportsService() *reportsservice.ReportsService {
	if d.reportsSvc == nil {
		d.reportsSvc = reportsservice.NewReportsService(
			d.Store().Reports(),
			d.Store().Tasks(),
			d.UsersClient(),
		)
	}
	return d.reportsSvc
}

func (d *diContainer) UsersClient() *usersserviceclient.UsersClient {
	if d.usersClient == nil {
		d.usersClient = usersserviceclient.NewUsersClient(d.cfg.HTTPClient.UsersClient)
	}
	return d.usersClient
}

func (d *diContainer) GitTasksClient() *gittasksclient.GitAuthClient {
	if d.gitTasksClient == nil {
		d.gitTasksClient = gittasksclient.NewGitAuthClient(d.cfg.HTTPClient.GitTasksClient)
	}
	return d.gitTasksClient
}

func (d *diContainer) TasksHandlers() *handlers.TasksHandlers {
	if d.tasksHandlers == nil {
		d.tasksHandlers = handlers.NewTasksHandlers(d.TasksService(), d.TagsService(), d.LanguagesService())
	}
	return d.tasksHandlers
}

func (d *diContainer) GitTasksHandler() *gittaskshandler.GitTasksHandler {
	if d.gitTasksHandler == nil {
		d.gitTasksHandler = gittaskshandler.NewGitTasksHandler(d.GitTasksService())
	}
	return d.gitTasksHandler
}

func (d *diContainer) ReportsHandler() *reportshandler.ReportsHandler {
	if d.reportsHandler == nil {
		d.reportsHandler = reportshandler.NewReportsHandler(d.ReportsService())
	}
	return d.reportsHandler
}
