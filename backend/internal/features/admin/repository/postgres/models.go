package postgres

type AdminUserModel struct {
	ID          int
	Version     int
	FullName    string
	Email       string
	Password    string
	PhoneNumber *string
	IsAdmin     bool
}

type MetricsModel struct {
	TotalUsers        int     `json:"total_users"`
	TotalTransactions int     `json:"total_transactions"`
	TotalBalance      float64 `json:"total_balance"`
}
