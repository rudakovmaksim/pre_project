package storage

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"project/internal/adapters/configs"
	"project/internal/entities"
	"project/logger"
)

const (
	typeParamSSLmode = "sslmode"
	typeDriverSQL    = "postgres"
)

type Storage struct {
	numberOfUnicTitles int
	db                 *sql.DB
}

func NewStorage(cfg *configs.Configs) (*Storage, error) {
	// if err := logger.GetLogger(); err != nil {
	// 	return nil, errors.Wrapf(entities.ErrInternal, "fail create logger, err: %v", err)
	// }

	if cfg.ConnectString == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "connection string data base is empty")
	}

	logger.Infologger.Info("connection to the database at: ", zap.String("connect string", cfg.ConnectString))

	db, err := sql.Open(typeDriverSQL, cfg.ConnectString)
	if err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "failed to connect to database: %v", err)
	}

	if err = db.Ping(); err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "failed to ping database: %v", err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) Store(ctx context.Context, coins []entities.Coin) error {
	logger.Infologger.Info("call func Store", zap.Any("coins", coins))

	query := `
	WITH inserted_title AS (
		INSERT INTO titles (title)
		VALUES ($1)
		ON CONFLICT (title) DO NOTHING
		RETURNING id
	),
	title_id_cte AS (
		SELECT id
		FROM inserted_title
		UNION ALL
		SELECT id
		FROM titles
		WHERE title = $1
	),
	inserted_date AS (
		INSERT INTO dates (event_date)
		VALUES (CURRENT_DATE)
		ON CONFLICT (event_date) DO NOTHING
		RETURNING id
	),
	day_id_cte AS (
		SELECT id
		FROM inserted_date
		UNION ALL
		SELECT id
		FROM dates
		WHERE event_date = CURRENT_DATE
	)
	INSERT INTO rates (title_id, day_id, cost, event_time)
	SELECT
		(SELECT id FROM title_id_cte),
		(SELECT id FROM day_id_cte),
		$2,
		CURRENT_TIME;`

	for _, coin := range coins {
		_, err := s.db.ExecContext(ctx, query, coin.Title, coin.Cost)
		if err != nil {
			return errors.Wrapf(entities.ErrInternal, "failed func Exec to insert data to database: %v", err)
		}
	}

	return nil
}

func (s *Storage) GetCoinsList(ctx context.Context) ([]string, error) {
	logger.Infologger.Info("call func GetCoinsList")

	query := `SELECT title FROM titles`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrapf(entities.ErrNotFound, "not found data in database: %v", err)
		}
		return nil, errors.Wrapf(entities.ErrInternal, "failed to execute query to database: %v", err)
	}
	defer rows.Close()

	titleList := make([]string, 0, s.numberOfUnicTitles)

	for rows.Next() {
		var title string
		if err = rows.Scan(&title); err != nil {
			return nil, errors.Wrapf(entities.ErrInternal, "failed to scan rows that were retrieved from the database: %v", err)
		}

		titleList = append(titleList, title)
	}

	s.numberOfUnicTitles = len(titleList)

	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "failed iteration rows: %v", err)
	}

	return titleList, nil
}

func (s *Storage) GetActualCoins(ctx context.Context, titles []string) ([]entities.Coin, error) {
	logger.Infologger.Info("call func GetActualCoins", zap.Any("titles", titles))

	// ($1::text[])
	query := `
	SELECT t.title, r.cost
	FROM titles t
	JOIN rates r ON t.id = r.title_id
	WHERE t.title = ANY($1)
	  AND (r.day_id, r.event_time) = (
		  SELECT r2.day_id, r2.event_time
		  FROM rates r2
		  WHERE r2.title_id = t.id
		  ORDER BY r2.day_id DESC, r2.event_time DESC
		  LIMIT 1
	  );`

	rows, errQuery := s.db.QueryContext(ctx, query, pq.Array(titles))
	if errQuery != nil {
		if errors.Is(errQuery, sql.ErrNoRows) {
			return nil, errors.Wrapf(entities.ErrNotFound, "not found data in database: %v", errQuery)
		}
		return nil, errors.Wrapf(entities.ErrInternal, "failed to execute query to database: %v", errQuery)
	}
	defer rows.Close()

	coinList := make([]entities.Coin, 0, len(titles))
	for rows.Next() {
		var (
			title string
			cost  float64
		)

		if errScan := rows.Scan(&title, &cost); errScan != nil {
			return nil, errors.Wrapf(entities.ErrInternal, "failed to scan rows that were retrieved from the database: %v", errScan)
		}

		coin, errCoin := entities.NewCoin(title, cost)
		if errCoin != nil {
			return nil, errors.Wrap(errCoin, "fail create new coin")
		}

		coinList = append(coinList, *coin)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "failed iteration rows: %v", err)
	}

	return coinList, nil
}

func (s *Storage) GetAggregateCoins(ctx context.Context, titles []string, valueParameter string) ([]entities.Coin, error) {
	logger.Infologger.Info("call func GetAggregateCoins", zap.Any("titles", titles), zap.String("valueParameter", valueParameter))

	if valueParameter == "max" {
		query := `
		SELECT t.title, r.cost
		FROM titles t
		JOIN rates r ON t.id = r.title_id
		WHERE t.title = ANY($1)
		  AND r.id = (
			  SELECT DISTINCT ON (title_id) id
			  FROM rates
			  WHERE title_id = t.id
			  ORDER BY title_id, day_id DESC, cost DESC
		  );`

		rows, errQuery := s.db.QueryContext(ctx, query, pq.Array(titles))
		if errQuery != nil {
			if errors.Is(errQuery, sql.ErrNoRows) {
				return nil, errors.Wrapf(entities.ErrNotFound, "not found data in database: %v", errQuery)
			}
			return nil, errors.Wrapf(entities.ErrInternal, "failed to execute query to database: %v", errQuery)
		}
		defer rows.Close()

		coinList := make([]entities.Coin, 0, len(titles))
		for rows.Next() {
			var (
				title string
				cost  float64
			)
			if errScan := rows.Scan(&title, &cost); errScan != nil {
				return nil, errors.Wrapf(entities.ErrInternal, "failed to scan rows that were retrieved from the database: %v", errScan)
			}

			coin, errCoin := entities.NewCoin(title, cost)
			if errCoin != nil {
				return nil, errors.Wrap(errCoin, "fail create new coin")
			}

			coinList = append(coinList, *coin)
		}

		if err := rows.Err(); err != nil {
			return nil, errors.Wrapf(entities.ErrInternal, "failed iteration rows: %v", err)
		}

		return coinList, nil
	}

	if valueParameter == "min" {
		query := `
		SELECT t.title, r.cost
		FROM titles t
		JOIN rates r ON t.id = r.title_id
		WHERE r.id = (
			SELECT DISTINCT ON (title_id) id
			FROM rates
			WHERE title_id = t.id
			ORDER BY title_id, day_id DESC, cost ASC
		)
		AND t.title = ANY($1);`

		rows, errQuery := s.db.QueryContext(ctx, query, pq.Array(titles))
		if errQuery != nil {
			if errors.Is(errQuery, sql.ErrNoRows) {
				return nil, errors.Wrapf(entities.ErrNotFound, "not found data in database: %v", errQuery)
			}
			return nil, errors.Wrapf(entities.ErrInternal, "failed to execute query to database: %v", errQuery)
		}
		defer rows.Close()

		coinList := make([]entities.Coin, 0, len(titles))
		for rows.Next() {
			var (
				title string
				cost  float64
			)

			if errScan := rows.Scan(&title, &cost); errScan != nil {
				return nil, errors.Wrapf(entities.ErrInternal, "failed to scan rows that were retrieved from the database: %v", errScan)
			}

			coin, errCoin := entities.NewCoin(title, cost)
			if errCoin != nil {
				return nil, errors.Wrap(errCoin, "fail create new coin")
			}

			coinList = append(coinList, *coin)
		}

		if err := rows.Err(); err != nil {
			return nil, errors.Wrapf(entities.ErrInternal, "failed iteration rows: %v", err)
		}

		return coinList, nil
	}

	if valueParameter == "average" {
		query := `
		SELECT t.title, ROUND(AVG(r.cost::numeric), 3)
		FROM titles t
		JOIN rates r ON t.id = r.title_id
		WHERE t.title = ANY($1)
		  AND r.day_id = (
			  SELECT MAX(day_id)
			  FROM rates
			  WHERE title_id = t.id
		  )
		GROUP BY t.title;`

		rows, errQuery := s.db.QueryContext(ctx, query, pq.Array(titles))
		if errQuery != nil {
			if errors.Is(errQuery, sql.ErrNoRows) {
				return nil, errors.Wrapf(entities.ErrNotFound, "not found data in database: %v", errQuery)
			}
			return nil, errors.Wrapf(entities.ErrInternal, "failed to execute query to database: %v", errQuery)
		}
		defer rows.Close()

		coinList := make([]entities.Coin, 0, len(titles))
		for rows.Next() {
			var (
				title string
				cost  float64
			)

			if errScan := rows.Scan(&title, &cost); errScan != nil {
				return nil, errors.Wrapf(entities.ErrInternal, "failed to scan rows that were retrieved from the database: %v", errScan)
			}

			coin, errCoin := entities.NewCoin(title, cost)
			if errCoin != nil {
				return nil, errors.Wrap(errCoin, "fail create new coin")
			}

			coinList = append(coinList, *coin)
		}

		if err := rows.Err(); err != nil {
			return nil, errors.Wrapf(entities.ErrInternal, "failed iteration rows: %v", err)
		}

		return coinList, nil
	}

	if valueParameter == "percent" {
		query := `
		WITH elapsed_time AS (
			SELECT 
				t.title, 
				r.cost::numeric AS cost
			FROM 
				titles t
			JOIN 
				rates r ON t.id = r.title_id
			WHERE 
				t.title = ANY($1)
				AND r.day_id = (
					SELECT MAX(day_id)
					FROM rates
					WHERE title_id = t.id
				)
				AND r.event_time >= (NOW() - INTERVAL '1 hour')::time 
				AND r.event_time < NOW()::time 
			ORDER BY 
				r.event_time DESC
			LIMIT 1
		),
		now_time AS (
			SELECT 
				t.title, 
				r.cost::numeric AS cost
			FROM 
				titles t
			JOIN 
				rates r ON t.id = r.title_id
			WHERE 
				t.title = ANY($1)
				AND (r.day_id, r.event_time) = (
					SELECT 
						day_id, event_time
					FROM 
						rates
					WHERE 
						title_id = t.id
					ORDER BY 
						day_id DESC, event_time DESC
					LIMIT 1
				)
		)
		SELECT 
			n.title, 
			n.cost AS current_cost,
			ROUND((n.cost - e.cost) / NULLIF(e.cost, 0) * 100, 2) AS percent_change
		FROM 
			now_time n
		JOIN 
			elapsed_time e ON n.title = e.title;`

		rows, errQuery := s.db.QueryContext(ctx, query, pq.Array(titles))
		if errQuery != nil {
			if errors.Is(errQuery, sql.ErrNoRows) {
				return nil, errors.Wrapf(entities.ErrNotFound, "not found data in database: %v", errQuery)
			}
			return nil, errors.Wrapf(entities.ErrInternal, "failed to execute query to database: %v", errQuery)
		}
		defer rows.Close()

		coinList := make([]entities.Coin, 0, len(titles))
		for rows.Next() {
			var (
				title   string
				cost    float64
				percent float32
			)

			if errScan := rows.Scan(&title, &cost, &percent); errScan != nil {
				return nil, errors.Wrapf(entities.ErrInternal, "failed to scan rows that were retrieved from the database: %v", errScan)
			}

			coin, errCoin := entities.NewCoin(title, cost)
			coin.RatePercentDelta = percent
			if errCoin != nil {
				return nil, errors.Wrap(errCoin, "fail create new coin")
			}

			coinList = append(coinList, *coin)
		}

		if err := rows.Err(); err != nil {
			return nil, errors.Wrapf(entities.ErrInternal, "failed iteration rows: %v", err)
		}

		return coinList, nil
	}

	return nil, errors.Wrap(entities.ErrInvalidAggregateParam, "incorrect parameter passed to the func GetAggregateCoins")
}
