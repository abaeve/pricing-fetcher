// Automatically generated by MockGen. DO NOT EDIT!
// Source: publisher.go

package fetching

import (
	gomock "github.com/golang/mock/gomock"
)

// Mock of OrderPublisher interface
type MockOrderPublisher struct {
	ctrl     *gomock.Controller
	recorder *_MockOrderPublisherRecorder
}

// Recorder for MockOrderPublisher (not exported)
type _MockOrderPublisherRecorder struct {
	mock *MockOrderPublisher
}

func NewMockOrderPublisher(ctrl *gomock.Controller) *MockOrderPublisher {
	mock := &MockOrderPublisher{ctrl: ctrl}
	mock.recorder = &_MockOrderPublisherRecorder{mock}
	return mock
}

func (_m *MockOrderPublisher) EXPECT() *_MockOrderPublisherRecorder {
	return _m.recorder
}

func (_m *MockOrderPublisher) PublishOrder(order *OrderPayload) {
	_m.ctrl.Call(_m, "PublishOrder", order)
}

func (_mr *_MockOrderPublisherRecorder) PublishOrder(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "PublishOrder", arg0)
}

func (_m *MockOrderPublisher) PublishStateBegin(regionId int32) {
	_m.ctrl.Call(_m, "PublishStateBegin", regionId)
}

func (_mr *_MockOrderPublisherRecorder) PublishStateBegin(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "PublishStateBegin", arg0)
}

func (_m *MockOrderPublisher) PublishStateEnd(regionId int32) {
	_m.ctrl.Call(_m, "PublishStateEnd", regionId)
}

func (_mr *_MockOrderPublisherRecorder) PublishStateEnd(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "PublishStateEnd", arg0)
}