package domain

import "time"

type Flow string

const (
	FlowTransaction Flow = "transaction"
	FlowBudget      Flow = "budget"
	FlowShopping    Flow = "shopping"
	FlowAIPlan      Flow = "ai_plan"
	FlowCurrency    Flow = "currency"
	FlowFilter      Flow = "finance_filter"
)

type DialogState struct {
	ChatID    int64
	UserID    int
	Flow      Flow
	Step      string
	Data      map[string]string
	MessageID int
	UpdatedAt time.Time
}

func (s *DialogState) Set(key, value string) {
	if s.Data == nil {
		s.Data = make(map[string]string)
	}
	s.Data[key] = value
}

func (s DialogState) Get(key string) string {
	if s.Data == nil {
		return ""
	}
	return s.Data[key]
}

const (
	StepTransactionType     = "transaction_type"
	StepTransactionAmount   = "transaction_amount"
	StepTransactionCategory = "transaction_category"
	StepTransactionNote     = "transaction_note"
	StepTransactionConfirm  = "transaction_confirm"

	StepBudgetCategory = "budget_category"
	StepBudgetAmount   = "budget_amount"
	StepBudgetPeriod   = "budget_period"
	StepBudgetConfirm  = "budget_confirm"

	StepShoppingTitle       = "shopping_title"
	StepShoppingDescription = "shopping_description"
	StepShoppingAmountNow   = "shopping_amount_now"
	StepShoppingAmountGoal  = "shopping_amount_goal"
	StepShoppingConfirm     = "shopping_confirm"
	StepShoppingImage       = "shopping_image"

	StepAIPlanText    = "ai_plan_text"
	StepAIPlanAddress = "ai_plan_address"
	StepAIPlanGeo     = "ai_plan_geo"

	StepCurrencyFrom   = "currency_from"
	StepCurrencyTo     = "currency_to"
	StepCurrencyAmount = "currency_amount"

	StepFilterDates = "filter_dates"
)

const (
	ActionIncome  = "income"
	ActionExpense = "expense"
	ActionConfirm = "confirm"
	ActionCancel  = "cancel"
	ActionSkip    = "skip"
)
