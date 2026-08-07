package mocks

import (
	"context"
	"reflect"

	"github.com/golang/mock/gomock"
)

type MockTx struct {
	ctrl     *gomock.Controller
	recorder *MockTxMockRecorder
}

type MockTxMockRecorder struct {
	mock *MockTx
}

func NewMockTx(ctrl *gomock.Controller) *MockTx {
	mock := &MockTx{ctrl: ctrl}
	mock.recorder = &MockTxMockRecorder{mock}
	return mock
}

func (m *MockTx) EXPECT() *MockTxMockRecorder {
	return m.recorder
}

func (m *MockTx) Commit(ctx context.Context) error {
	ret := m.ctrl.Call(m, "Commit", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockTxMockRecorder) Commit(ctx interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(
		mr.mock,
		"Commit",
		reflect.TypeOf((*MockTx)(nil).Commit),
		ctx,
	)
}

func (m *MockTx) Rollback(ctx context.Context) error {
	ret := m.ctrl.Call(m, "Rollback", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockTxMockRecorder) Rollback(ctx interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(
		mr.mock,
		"Rollback",
		reflect.TypeOf((*MockTx)(nil).Rollback),
		ctx,
	)
}
