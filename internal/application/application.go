package application

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"project/internal/adapters/configs"
	"project/internal/adapters/provider"
	"project/internal/adapters/storage"
	"project/internal/cases"
	"project/internal/entities"
	"project/internal/port"
	"project/logger"
)

type App struct {
	service *cases.Service
	server  *port.Server
}

func NewApp() (*App, error) {
	if err := logger.GetLogger(); err != nil {
		logger.Infologger.Error("error create logger", zap.Error(err))
	}

	cfg, err := configs.LoadConfig()
	if err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "error load configs: %v", err)
	}

	client, err := provider.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "error create client: %v", err)
	}

	storage, err := storage.NewStorage(cfg)
	if err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "error create storage: %v", err)
	}

	service, err := cases.NewService(client, storage, cfg)
	if err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "error create service: %v", err)
	}

	server, err := port.NewServer(service, cfg)
	if err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "error create server: %v", err)
	}

	return &App{
		service: service,
		server:  server,
	}, nil
}

func (a *App) Run() error {
	logger.Infologger.Info("starting application")

	errChan := make(chan error, 1)

	go func() {
		if err := startUpdateRates(a.service); err != nil {
			errChan <- err
		}
		errChan <- nil
	}()

	err := <-errChan

	if err != nil {
		return errors.Wrapf(entities.ErrInternal, "starting the cron job: %v", err)
	}

	if err = a.server.Start(); err != nil {
		return errors.Wrapf(entities.ErrInternal, "starting the data base: %v", err)
	}

	return nil
}
