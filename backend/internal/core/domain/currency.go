package domain

import (
	"fmt"
	"time"
)

type Rate struct {
	Base      string             `json:"base"`
	Rates     map[string]float64 `json:"rates"`
	Timestamp time.Time          `json:"timestamp"`
}

type Conversion struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Amount    float64   `json:"amount"`
	Result    float64   `json:"result"`
	Rate      float64   `json:"rate"`
	Timestamp time.Time `json:"timestamp"`
}

type TransactionUSD struct {
	TransactionID    int       `json:"transaction_id"`
	AmountUSD        float64   `json:"amount_usd"`
	OriginalAmount   float64   `json:"original_amount"`
	OriginalCurrency string    `json:"original_currency"`
	ConvertedAt      time.Time `json:"converted_at"`
}

func (r *Rate) Validate() error {
	if r.Base == "" || len(r.Base) <= 0 {
		return fmt.Errorf("invalid base len <= 0 or len = ''")
	}
	if r.Rates == nil || len(r.Rates) <= 0 {
		return fmt.Errorf("invalid rates len <= 0 or len = ''")
	}
	return nil
}

func (c *Conversion) Validate() error {
	if c.From == "" || len(c.From) <= 0 {
		return fmt.Errorf("invalid from len < 0 or len = ''")
	}
	if c.To == "" || len(c.To) <= 0 {
		return fmt.Errorf("invalid to len < 0 or len = ''")
	}
	if c.Amount < 0 {
		return fmt.Errorf("invalid amount < 0")
	}
	if c.Result < 0 {
		return fmt.Errorf("invalid result < 0")
	}
	if c.Rate < 0 {
		return fmt.Errorf("invalid rate < 0")
	}
	return nil
}

func (r *TransactionUSD) Validate() error {
	if r.TransactionID <= 0 {
		return fmt.Errorf("invalid transaction id <= 0")
	}
	if r.AmountUSD <= 0 {
		return fmt.Errorf("invalid amount <= 0")
	}
	if r.OriginalAmount <= 0 {
		return fmt.Errorf("invalid original_amount <= 0")
	}
	if r.OriginalCurrency == "" || len(r.OriginalCurrency) <= 0 {
		return fmt.Errorf("invalid original_currency is empty")
	}
	return nil
}
