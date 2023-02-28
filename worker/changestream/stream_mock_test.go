// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/worker/changestream (interfaces: ChangeStream,DBGetter,EventQueueWorker,EventQueue,FileNotifyWatcher)

// Package changestream is a generated GoMock package.
package changestream

import (
	sql "database/sql"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	changestream "github.com/juju/juju/core/changestream"
)

// MockChangeStream is a mock of ChangeStream interface.
type MockChangeStream struct {
	ctrl     *gomock.Controller
	recorder *MockChangeStreamMockRecorder
}

// MockChangeStreamMockRecorder is the mock recorder for MockChangeStream.
type MockChangeStreamMockRecorder struct {
	mock *MockChangeStream
}

// NewMockChangeStream creates a new mock instance.
func NewMockChangeStream(ctrl *gomock.Controller) *MockChangeStream {
	mock := &MockChangeStream{ctrl: ctrl}
	mock.recorder = &MockChangeStreamMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockChangeStream) EXPECT() *MockChangeStreamMockRecorder {
	return m.recorder
}

// EventQueue mocks base method.
func (m *MockChangeStream) EventQueue(arg0 string) (EventQueue, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EventQueue", arg0)
	ret0, _ := ret[0].(EventQueue)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EventQueue indicates an expected call of EventQueue.
func (mr *MockChangeStreamMockRecorder) EventQueue(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EventQueue", reflect.TypeOf((*MockChangeStream)(nil).EventQueue), arg0)
}

// MockDBGetter is a mock of DBGetter interface.
type MockDBGetter struct {
	ctrl     *gomock.Controller
	recorder *MockDBGetterMockRecorder
}

// MockDBGetterMockRecorder is the mock recorder for MockDBGetter.
type MockDBGetterMockRecorder struct {
	mock *MockDBGetter
}

// NewMockDBGetter creates a new mock instance.
func NewMockDBGetter(ctrl *gomock.Controller) *MockDBGetter {
	mock := &MockDBGetter{ctrl: ctrl}
	mock.recorder = &MockDBGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDBGetter) EXPECT() *MockDBGetterMockRecorder {
	return m.recorder
}

// GetDB mocks base method.
func (m *MockDBGetter) GetDB(arg0 string) (*sql.DB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDB", arg0)
	ret0, _ := ret[0].(*sql.DB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDB indicates an expected call of GetDB.
func (mr *MockDBGetterMockRecorder) GetDB(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDB", reflect.TypeOf((*MockDBGetter)(nil).GetDB), arg0)
}

// MockEventQueueWorker is a mock of EventQueueWorker interface.
type MockEventQueueWorker struct {
	ctrl     *gomock.Controller
	recorder *MockEventQueueWorkerMockRecorder
}

// MockEventQueueWorkerMockRecorder is the mock recorder for MockEventQueueWorker.
type MockEventQueueWorkerMockRecorder struct {
	mock *MockEventQueueWorker
}

// NewMockEventQueueWorker creates a new mock instance.
func NewMockEventQueueWorker(ctrl *gomock.Controller) *MockEventQueueWorker {
	mock := &MockEventQueueWorker{ctrl: ctrl}
	mock.recorder = &MockEventQueueWorkerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEventQueueWorker) EXPECT() *MockEventQueueWorkerMockRecorder {
	return m.recorder
}

// EventQueue mocks base method.
func (m *MockEventQueueWorker) EventQueue() EventQueue {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EventQueue")
	ret0, _ := ret[0].(EventQueue)
	return ret0
}

// EventQueue indicates an expected call of EventQueue.
func (mr *MockEventQueueWorkerMockRecorder) EventQueue() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EventQueue", reflect.TypeOf((*MockEventQueueWorker)(nil).EventQueue))
}

// Kill mocks base method.
func (m *MockEventQueueWorker) Kill() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Kill")
}

// Kill indicates an expected call of Kill.
func (mr *MockEventQueueWorkerMockRecorder) Kill() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Kill", reflect.TypeOf((*MockEventQueueWorker)(nil).Kill))
}

// Wait mocks base method.
func (m *MockEventQueueWorker) Wait() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Wait")
	ret0, _ := ret[0].(error)
	return ret0
}

// Wait indicates an expected call of Wait.
func (mr *MockEventQueueWorkerMockRecorder) Wait() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Wait", reflect.TypeOf((*MockEventQueueWorker)(nil).Wait))
}

// MockEventQueue is a mock of EventQueue interface.
type MockEventQueue struct {
	ctrl     *gomock.Controller
	recorder *MockEventQueueMockRecorder
}

// MockEventQueueMockRecorder is the mock recorder for MockEventQueue.
type MockEventQueueMockRecorder struct {
	mock *MockEventQueue
}

// NewMockEventQueue creates a new mock instance.
func NewMockEventQueue(ctrl *gomock.Controller) *MockEventQueue {
	mock := &MockEventQueue{ctrl: ctrl}
	mock.recorder = &MockEventQueueMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEventQueue) EXPECT() *MockEventQueueMockRecorder {
	return m.recorder
}

// Subscribe mocks base method.
func (m *MockEventQueue) Subscribe(arg0 func(changestream.ChangeEvent), arg1 ...changestream.SubscriptionOption) (changestream.Subscription, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Subscribe", varargs...)
	ret0, _ := ret[0].(changestream.Subscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockEventQueueMockRecorder) Subscribe(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockEventQueue)(nil).Subscribe), varargs...)
}

// MockFileNotifyWatcher is a mock of FileNotifyWatcher interface.
type MockFileNotifyWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockFileNotifyWatcherMockRecorder
}

// MockFileNotifyWatcherMockRecorder is the mock recorder for MockFileNotifyWatcher.
type MockFileNotifyWatcherMockRecorder struct {
	mock *MockFileNotifyWatcher
}

// NewMockFileNotifyWatcher creates a new mock instance.
func NewMockFileNotifyWatcher(ctrl *gomock.Controller) *MockFileNotifyWatcher {
	mock := &MockFileNotifyWatcher{ctrl: ctrl}
	mock.recorder = &MockFileNotifyWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFileNotifyWatcher) EXPECT() *MockFileNotifyWatcherMockRecorder {
	return m.recorder
}

// Changes mocks base method.
func (m *MockFileNotifyWatcher) Changes(arg0 string) (<-chan bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Changes", arg0)
	ret0, _ := ret[0].(<-chan bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Changes indicates an expected call of Changes.
func (mr *MockFileNotifyWatcherMockRecorder) Changes(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Changes", reflect.TypeOf((*MockFileNotifyWatcher)(nil).Changes), arg0)
}
