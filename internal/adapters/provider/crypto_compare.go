package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"project/internal/adapters/configs"
	"project/internal/entities"
	"project/logger"
)

const (
	typeParamFromSymbols = "fsyms"
	typeParamToSymbols   = "tsyms"
	typeParamAPIKey      = "api_key"
)

type cryptoData struct {
	Raw map[string]map[string]struct {
		Price float64 `json:"PRICE"`
	} `json:"RAW"`
}

type Client struct {
	key           string
	exchangeRates string
	baseURL       *url.URL
	client        http.Client
}

func NewClient(cfg *configs.Configs) (*Client, error) {
	if cfg.URL == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "url is empty")
	}

	baseURL, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "fail create url, err: %v", err)
	}

	// if err := logger.GetLogger(); err != nil {
	// 	return nil, errors.Wrapf(entities.ErrInternal, "fail create logger, err: %v", err)
	// }

	return &Client{
		key:           cfg.Key,
		exchangeRates: cfg.ExchangeRates,
		baseURL:       baseURL,
		client:        http.Client{},
	}, nil
}

func (c *Client) GetActualRates(ctx context.Context, titles []string) ([]entities.Coin, error) {
	logger.Infologger.Info("request to crypto compare")

	query := c.baseURL.Query()
	query.Set(typeParamFromSymbols, strings.Join(titles, ","))
	query.Set(typeParamToSymbols, c.exchangeRates)
	query.Set(typeParamAPIKey, c.key)

	c.baseURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL.String(), nil)
	if err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "failed to create HTTP request: %v", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "failed to execute HTTP request: %v", err)
	}
	defer resp.Body.Close()
	// правильно ли я обработал ошибку при получении не удачного статуса запроса
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Wrap(entities.ErrInternal, "getting failed request status")
	}

	data := cryptoData{
		Raw: make(map[string]map[string]struct {
			Price float64 `json:"PRICE"`
		}, len(titles)),
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "fail decoderation response: %v", err)
	}

	for _, title := range titles {
		if _, ok := data.Raw[title]; !ok {
			return nil, errors.Wrap(entities.ErrNotFound, "not found tites by request")
		}
	}

	coins := make([]entities.Coin, 0, len(titles))

	for title, coin := range data.Raw {
		cost := coin[c.exchangeRates]

		tempCoin, err := entities.NewCoin(title, cost.Price)
		if err != nil {
			return nil, errors.Wrap(err, "fail create new coin")
		}

		coins = append(coins, *tempCoin)
	}

	logger.Infologger.Info("returned response data: ", zap.Any("coins", coins))

	return coins, nil
}
