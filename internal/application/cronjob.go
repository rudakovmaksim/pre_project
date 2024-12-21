package application

import (
	"context"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"project/internal/cases"
	"project/internal/entities"
	"project/logger"
)

const updateCronFuncRates = "cronupdate"

func startUpdateRates(service *cases.Service) error {
	logger.Infologger.Info("start cron job func")
	ctx := context.TODO()

	cronJob := cron.New()

	_, errJob := cronJob.AddFunc("*/1 * * * *", func() {
		logger.Infologger.Info("CALLING CRON JOB FOR UPDATE")
		errServ := service.ActualizeRates(ctx, updateCronFuncRates)
		if errServ != nil {
			logger.Infologger.Warn("error call func ActualizeRates in cron job: ", zap.String("errServ", errServ.Error()))
		}
		logger.Infologger.Info("END OF CALL CRON JOB FOR UPDATE")
	})

	if errJob != nil {
		return errors.Wrapf(entities.ErrInternal, "error working cron job: %v", errJob)
	}

	cronJob.Start()

	return nil
}
