package logger

import (
	"project/internal/entities"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	once       sync.Once
	Infologger *zap.Logger
)

func GetLogger() error {
	var err error

	once.Do(func() {
		Infologger, err = zap.NewDevelopment()
		if err != nil {
			err = errors.Wrapf(entities.ErrInternal, "error create logger: %v", err)
		}
	})

	return err
}

func CloseLogger() error {
	if err := Infologger.Sync(); err != nil {
		return errors.Wrapf(entities.ErrInternal, "error close logger: %v", err)
	}
	return nil
}
