package testutil

import (
	"context"
	"reflect"

	"github.com/golang/mock/gomock"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

// MockMsgServer is a mock of MsgServer interface.
type MockMsgServer struct {
	ctrl     *gomock.Controller
	recorder *MockMsgServerMockRecorder
}

// MockMsgServerMockRecorder is the mock recorder for MockMsgServer.
type MockMsgServerMockRecorder struct {
	mock *MockMsgServer
}

// NewMockMsgServer creates a new mock instance.
func NewMockMsgServer(ctrl *gomock.Controller) *MockMsgServer {
	mock := &MockMsgServer{ctrl: ctrl}
	mock.recorder = &MockMsgServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMsgServer) EXPECT() *MockMsgServerMockRecorder {
	return m.recorder
}

// UpdateParams mocks base method.
func (m *MockMsgServer) UpdateParams(ctx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateParams", ctx, req)
	ret0, _ := ret[0].(*types.MsgUpdateParamsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateParams indicates an expected call of UpdateParams.
func (mr *MockMsgServerMockRecorder) UpdateParams(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateParams", reflect.TypeOf((*MockMsgServer)(nil).UpdateParams), ctx, req)
}

// MockQueryServer is a mock of QueryServer interface.
type MockQueryServer struct {
	ctrl     *gomock.Controller
	recorder *MockQueryServerMockRecorder
}

// MockQueryServerMockRecorder is the mock recorder for MockQueryServer.
type MockQueryServerMockRecorder struct {
	mock *MockQueryServer
}

// NewMockQueryServer creates a new mock instance.
func NewMockQueryServer(ctrl *gomock.Controller) *MockQueryServer {
	mock := &MockQueryServer{ctrl: ctrl}
	mock.recorder = &MockQueryServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQueryServer) EXPECT() *MockQueryServerMockRecorder {
	return m.recorder
}

// Params mocks base method.
func (m *MockQueryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Params", ctx, req)
	ret0, _ := ret[0].(*types.QueryParamsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Params indicates an expected call of Params.
func (mr *MockQueryServerMockRecorder) Params(ctx, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Params", reflect.TypeOf((*MockQueryServer)(nil).Params), ctx, req)
}
