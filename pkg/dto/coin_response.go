package dto

type CoinResponse struct {
	Title            string  `json:"title"`
	Cost             float64 `json:"cost"`
	RatePercentDelta float32 `json:"rate_percent_delta"`
}
