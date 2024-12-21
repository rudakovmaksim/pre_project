package port

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"project/internal/adapters/configs"
	"project/internal/entities"
	"project/logger"
	"project/pkg/dto"

	_ "project/docs" // Swagger documentation

	httpSwagger "github.com/swaggo/http-swagger"
)

// http://localhost:8080/swagger/index.html

// /rates/last?titles=BTC,ETH
// localhost:8080

type Server struct {
	service Service
	router  *chi.Mux
	address string
}

func (s *Server) setupRouter() {
	s.router = chi.NewRouter()

	s.router.Get("/rates/last", s.getLastRates)
	s.router.Get("/rates/{typeinfo}", s.getAvgRates)
	s.router.Get("/rates/{typeinfo}", s.getAvgRates)
	s.router.Get("/rates/{typeinfo}", s.getAvgRates)
	s.router.Get("/rates/{typeinfo}", s.getAvgRates)

	s.router.Get("/swagger/*", httpSwagger.WrapHandler)
}

func NewServer(service Service, cfg *configs.Configs) (*Server, error) {
	if service == nil || service == Service(nil) {
		return nil, errors.Wrap(entities.ErrInvalidParam, "service is nil")
	}

	// if err := logger.GetLogger(); err != nil {
	// 	return nil, errors.Wrapf(entities.ErrInternal, "fail create logger, err: %v", err)
	// }

	if cfg.Address == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "address sarver is empty")
	}

	serv := &Server{
		service: service,
		address: cfg.Address,
	}

	serv.setupRouter()

	return serv, nil
}

func (s *Server) Start() error {
	server := &http.Server{
		Addr:         s.address,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Infologger.Info("running server on the host: ", zap.String("address", server.Addr))

	if err := server.ListenAndServe(); err != nil {
		return errors.Wrapf(entities.ErrInternal, "error port startup: %v", err)
	}

	return nil
}

// @Summary Get Last Rates
// @Description Получить последние курсы криптовалют по указанным названиям.
// @Tags Rates
// @Accept json
// @Produce json
// @Param titles query string true "Список названий криптовалют через запятую (например, BTC,ETH)"
// @Success 200 {array} coinDTO "Список последних курсов криптовалют"
// @Failure 400 {string} string "Bad Request - отсутствует параметр titles"
// @Failure 404 {string} string "Not Found - данные не найдены"
// @Failure 500 {string} string "Internal Server Error - ошибка на сервере"
// @Router /rates/last [get]
func (s *Server) getLastRates(w http.ResponseWriter, req *http.Request) {
	logger.Infologger.Info("request func getLastRates")
	ctx := req.Context()

	titlesParam := req.URL.Query().Get("titles")

	if titlesParam == "" {
		http.Error(w, "titles parameter is required", http.StatusBadRequest)
		return
	}

	titles := strings.Split(titlesParam, ",")

	coins, err := s.service.GetLastRates(ctx, titles)
	if err != nil {
		if errors.As(err, &entities.ErrNotFound) {
			logger.Infologger.Info("not found data when query get last rates")
			http.Error(w, "Not found data", http.StatusNotFound)
			return
		}
		logger.Infologger.Error("error call func GetLastRates when query get last rates", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	coinsTransfer := make([]dto.CoinResponse, len(coins))
	for i, coin := range coins {
		coinsTransfer[i] = dto.CoinResponse{
			Title:            coin.Title,
			Cost:             coin.Cost,
			RatePercentDelta: coin.RatePercentDelta,
		}
	}

	js, err := json.Marshal(coinsTransfer)
	if err != nil {
		logger.Infologger.Error("error serialization a json when query get last rates", zap.Error(err))
		http.Error(w, "failed serialization a json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	numberByte, err := w.Write(js)
	if err != nil {
		if numberByte == len(js) {
			logger.Infologger.Error("client connection error, but all data was sent")
			return
		}
		logger.Infologger.Error("client connection error")
		return
	}

	logger.Infologger.Info("completed response to client info last", zap.Any("titles", titles), zap.Any("coins", coins))
}

// @Summary Get Average Rates
// @Description Получить агрегированные данные курсов криптовалют (максимум, минимум, среднее, процентное изменение).
// @Tags Rates
// @Accept json
// @Produce json
// @Param typeinfo path string true "Тип агрегации (max, min, avg, percent)"
// @Param titles query string true "Список названий криптовалют через запятую (например, BTC,ETH)"
// @Success 200 {array} coinDTO "Список агрегированных данных курсов криптовалют"
// @Failure 400 {string} string "Bad Request - отсутствует параметр titles или typeinfo"
// @Failure 404 {string} string "Not Found - данные не найдены"
// @Failure 500 {string} string "Internal Server Error - ошибка на сервере"
// @Router /rates/{typeinfo} [get]
func (s *Server) getAvgRates(w http.ResponseWriter, req *http.Request) {
	logger.Infologger.Info("request func getAvgReates")
	ctx := req.Context()

	typeInfo := chi.URLParam(req, "typeinfo")
	titlesParam := req.URL.Query().Get("titles")

	if typeInfo == "" && titlesParam == "" {
		http.Error(w, "titles parameter is required", http.StatusBadRequest)
		return
	}

	titles := strings.Split(titlesParam, ",")

	var (
		err   error
		coins []entities.Coin
	)

	switch typeInfo {
	case "max":
		coins, err = s.service.GetMaxRates(ctx, titles)

	case "min":
		coins, err = s.service.GetMinRates(ctx, titles)

	case "avg":
		coins, err = s.service.GetAvgRates(ctx, titles)

	case "percent":
		coins, err = s.service.GetPercentRates(ctx, titles)

	default:
		logger.Infologger.Info("completed response to parametres: ", zap.String("type info", typeInfo), zap.Any("type titles", titles))
		http.Error(w, "parameters is not correct", http.StatusNotFound)

		return
	}

	if err != nil {
		if errors.As(err, &entities.ErrNotFound) {
			logger.Infologger.Info("not found data when query get average rates")
			http.Error(w, "Not found data", http.StatusNotFound)
			return
		}
		logger.Infologger.Error("error when query get average rates", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	coinsTransfer := make([]dto.CoinResponse, len(titles))
	for i, coin := range coins {
		coinsTransfer[i] = dto.CoinResponse{
			Title:            coin.Title,
			Cost:             coin.Cost,
			RatePercentDelta: coin.RatePercentDelta,
		}
	}

	js, err := json.Marshal(coinsTransfer)
	if err != nil {
		http.Error(w, "failed serialization a json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	numberByte, err := w.Write(js)
	if err != nil {
		if numberByte == len(js) {
			logger.Infologger.Error("client connection error, but all data was sent")
			return
		}
		logger.Infologger.Error("client connection error")
		return
	}

	logger.Infologger.Info("completed response to client", zap.Any("type info", typeInfo), zap.Any("titles", titles), zap.Any("coins", coins))
}
