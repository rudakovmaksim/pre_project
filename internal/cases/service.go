package cases

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"project/internal/adapters/configs"
	"project/internal/entities"
	"project/logger"
)

const (
	getAggregateMax     = "max"
	getAggregateMin     = "min"
	getAggregateAverage = "average"
	getAggregatePercent = "percent"
)

const updateNotExistRates = "lastupdate"

type Service struct {
	cryptoProvider         CryptoProvider
	storage                Storage
	ourActualTitles        []string
	ourCheckNotExistTitles []string
}

func NewService(cryptoProvider CryptoProvider, storage Storage, cfg *configs.Configs) (*Service, error) {
	if cryptoProvider == nil || cryptoProvider == CryptoProvider(nil) {
		return nil, errors.Wrap(entities.ErrInvalidParam, "crypto provider is nil")
	}

	if storage == nil || storage == Storage(nil) {
		return nil, errors.Wrap(entities.ErrInvalidParam, "storage is nil")
	}

	if err := logger.GetLogger(); err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "fail create logger, err: %v", err)
	}

	return &Service{
		cryptoProvider:  cryptoProvider,
		storage:         storage,
		ourActualTitles: cfg.ActualTitles,
	}, nil
}

func (s *Service) ActualizeRates(ctx context.Context, scriptTypeUpdateRates string) error {
	if scriptTypeUpdateRates == "cronupdate" {
		logger.Infologger.Info("call func ActualizeRates", zap.String("scriptType", scriptTypeUpdateRates))

		var (
			requestTitles []string
			err           error
		)

		if len(s.ourActualTitles) != 0 {
			requestTitles = s.ourActualTitles
		} else {
			requestTitles, err = s.storage.GetCoinsList(ctx)
			if err != nil {
				return errors.Wrap(err, "failed requesting to current list coins from storage")
			}
		}

		actualRates, err := s.cryptoProvider.GetActualRates(ctx, requestTitles)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve actual rates from cryptoProvider")
		}

		err = s.storage.Store(ctx, actualRates)
		if err != nil {
			return errors.Wrap(err, "error saving actual rates to storage")
		}

		return nil
	}

	if scriptTypeUpdateRates == "lastupdate" {
		logger.Infologger.Info("call func ActualizeRates", zap.String("scriptType", scriptTypeUpdateRates))

		missingRates, err := s.cryptoProvider.GetActualRates(ctx, s.ourCheckNotExistTitles)
		if err != nil {
			if errors.As(err, &entities.ErrNotFound) {
				return errors.Wrap(entities.ErrNotFound, "error returned, not found rates in func GetActualRates")
			}
			return errors.Wrap(err, "error call func GetActualRates")
		}

		err = s.storage.Store(ctx, missingRates)
		if err != nil {
			return errors.Wrap(err, "error saving actual rates to storage")
		}

		return nil
	}

	return errors.Wrap(entities.ErrInvalidScript, "parameter script update rates is incorrect")
}

func (s *Service) GetLastRates(ctx context.Context, titles []string) ([]entities.Coin, error) {
	logger.Infologger.Info("call func GetLastRates", zap.Any("titles", titles))

	success, err := s.checkExistTitles(ctx, titles)
	if err != nil {
		return nil, errors.Wrap(err, "failed check exist titles")
	}

	if success {
		actualCoins, errCoin := s.storage.GetActualCoins(ctx, titles)
		if errCoin != nil {
			return nil, errors.Wrap(errCoin, "failed to retrieve actual coins from storage")
		}

		return actualCoins, nil
	}

	err = s.ActualizeRates(ctx, updateNotExistRates)
	if err != nil {
		return nil, errors.Wrap(err, "failed call ActualizeRates function")
	}

	s.ourActualTitles = append(s.ourActualTitles, s.ourCheckNotExistTitles...)

	actualCoins, err := s.storage.GetActualCoins(ctx, titles)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve actual coins from storage")
	}

	return actualCoins, nil
}

func (s *Service) GetMaxRates(ctx context.Context, titles []string) ([]entities.Coin, error) {
	logger.Infologger.Info("call func GetMaxRates", zap.Any("titles", titles))

	success, err := s.checkExistTitles(ctx, titles)
	if err != nil {
		return nil, errors.Wrap(err, "failed check exist titles")
	}

	if success {
		actualCoins, err := s.storage.GetAggregateCoins(ctx, titles, getAggregateMax)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve aggtegate max coin from storage")
		}

		return actualCoins, nil
	}

	return nil, errors.Wrap(entities.ErrNotFound, "don't have information max coin for the day")
}

func (s *Service) GetMinRates(ctx context.Context, titles []string) ([]entities.Coin, error) {
	logger.Infologger.Info("call func GetMinRates", zap.Any("titles", titles))

	success, err := s.checkExistTitles(ctx, titles)
	if err != nil {
		return nil, errors.Wrap(err, "failed check exist titles")
	}

	if success {
		actualCoins, err := s.storage.GetAggregateCoins(ctx, titles, getAggregateMin)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve aggtegate min coin from storage")
		}

		return actualCoins, nil
	}

	return nil, errors.Wrap(entities.ErrNotFound, "don't have information about min coin for the day")
}

func (s *Service) GetAvgRates(ctx context.Context, titles []string) ([]entities.Coin, error) {
	logger.Infologger.Info("call func GetAvgRates", zap.Any("titles", titles))

	success, err := s.checkExistTitles(ctx, titles)
	if err != nil {
		return nil, errors.Wrap(err, "failed check exist titles")
	}

	if success {
		actualCoins, err := s.storage.GetAggregateCoins(ctx, titles, getAggregateAverage)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve aggtegate average coin from storage")
		}

		return actualCoins, nil
	}

	return nil, errors.Wrap(entities.ErrNotFound, "don't have information about average coin for the day")
}

func (s *Service) GetPercentRates(ctx context.Context, titles []string) ([]entities.Coin, error) {
	logger.Infologger.Info("call func GetPercentRates", zap.Any("titles", titles))

	success, err := s.checkExistTitles(ctx, titles)
	if err != nil {
		return nil, errors.Wrap(err, "failed check exist titles")
	}

	if success {
		actualCoins, err := s.storage.GetAggregateCoins(ctx, titles, getAggregatePercent)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retrieve percentage changes coin from storage")
		}

		return actualCoins, nil
	}

	return nil, errors.Wrap(entities.ErrNotFound, "don't have information about percentage changes coin from storage")
}

func (s *Service) checkExistTitles(ctx context.Context, titles []string) (bool, error) {
	existListCoins, err := s.storage.GetCoinsList(ctx)
	if err != nil {
		return false, errors.Wrap(err, "failed to requesting a list titles from storage")
	}

	existMap := make(map[string]struct{}, len(existListCoins))
	for _, title := range existListCoins {
		existMap[title] = struct{}{}
	}

	if success := s.processCheckExist(existMap, titles); success {
		return true, nil
	}

	return false, nil
}

func (s *Service) processCheckExist(existListTitles map[string]struct{}, listRequestTitles []string) bool {
	var (
		missingTitles []string
		exists        = true
	)

	s.ourCheckNotExistTitles = make([]string, 0, len(listRequestTitles))

	for _, value := range listRequestTitles {
		if _, ok := existListTitles[value]; !ok {
			missingTitles = append(missingTitles, value)

			exists = false
		}
	}

	if !exists {
		s.ourCheckNotExistTitles = append(s.ourCheckNotExistTitles, missingTitles...)
	}

	return exists
}
