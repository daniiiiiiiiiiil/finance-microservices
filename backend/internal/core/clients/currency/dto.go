package currency

type GetRatesRequest struct {
	Base string
}

type ConvertRequest struct {
	From   string
	To     string
	Amount float64
}

type ConvertBatchRequest struct {
	From   string
	To     []string
	Amount float64
}
