package domain

type SagaType string

const (
	SagaTypeDeleteUser     SagaType = "delete_user"
	SagaTypeDeleteShopping SagaType = "delete_shopping"
	SagaTypeTransferMoney  SagaType = "transfer_money"
	SagaTypeRegisterUser   SagaType = "register_user"
)

func (st SagaType) String() string {
	return string(st)
}

func (st SagaType) IsValid() bool {
	switch st {
	case SagaTypeDeleteUser,
		SagaTypeDeleteShopping,
		SagaTypeTransferMoney,
		SagaTypeRegisterUser:
		return true
	}
	return false
}
