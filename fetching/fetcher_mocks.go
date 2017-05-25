// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/abaeve/pricing-fetcher/fetching (interfaces: RegionsFetcher,OrderFetcher,HistoryFetcher,OrderPublisher)

package fetching

import (
	v1 "github.com/antihax/goesi/v1"
	gomock "github.com/golang/mock/gomock"
	http "net/http"
)

// Mock of RegionsFetcher interface
type MockRegionsFetcher struct {
	ctrl     *gomock.Controller
	recorder *_MockRegionsFetcherRecorder
}

// Recorder for MockRegionsFetcher (not exported)
type _MockRegionsFetcherRecorder struct {
	mock *MockRegionsFetcher
}

func NewMockRegionsFetcher(ctrl *gomock.Controller) *MockRegionsFetcher {
	mock := &MockRegionsFetcher{ctrl: ctrl}
	mock.recorder = &_MockRegionsFetcherRecorder{mock}
	return mock
}

func (_m *MockRegionsFetcher) EXPECT() *_MockRegionsFetcherRecorder {
	return _m.recorder
}

func (_m *MockRegionsFetcher) GetUniverseRegions(_param0 map[string]interface{}) ([]int32, *http.Response, error) {
	ret := _m.ctrl.Call(_m, "GetUniverseRegions", _param0)
	ret0, _ := ret[0].([]int32)
	ret1, _ := ret[1].(*http.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (_mr *_MockRegionsFetcherRecorder) GetUniverseRegions(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetUniverseRegions", arg0)
}

func (_m *MockRegionsFetcher) GetUniverseRegionsRegionId(_param0 int32, _param1 map[string]interface{}) (v1.GetUniverseRegionsRegionIdOk, *http.Response, error) {
	ret := _m.ctrl.Call(_m, "GetUniverseRegionsRegionId", _param0, _param1)
	ret0, _ := ret[0].(v1.GetUniverseRegionsRegionIdOk)
	ret1, _ := ret[1].(*http.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (_mr *_MockRegionsFetcherRecorder) GetUniverseRegionsRegionId(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetUniverseRegionsRegionId", arg0, arg1)
}

// Mock of OrderFetcher interface
type MockOrderFetcher struct {
	ctrl     *gomock.Controller
	recorder *_MockOrderFetcherRecorder
}

// Recorder for MockOrderFetcher (not exported)
type _MockOrderFetcherRecorder struct {
	mock *MockOrderFetcher
}

func NewMockOrderFetcher(ctrl *gomock.Controller) *MockOrderFetcher {
	mock := &MockOrderFetcher{ctrl: ctrl}
	mock.recorder = &_MockOrderFetcherRecorder{mock}
	return mock
}

func (_m *MockOrderFetcher) EXPECT() *_MockOrderFetcherRecorder {
	return _m.recorder
}

func (_m *MockOrderFetcher) GetMarketsRegionIdOrders(_param0 string, _param1 int32, _param2 map[string]interface{}) ([]v1.GetMarketsRegionIdOrders200Ok, *http.Response, error) {
	ret := _m.ctrl.Call(_m, "GetMarketsRegionIdOrders", _param0, _param1, _param2)
	ret0, _ := ret[0].([]v1.GetMarketsRegionIdOrders200Ok)
	ret1, _ := ret[1].(*http.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (_mr *_MockOrderFetcherRecorder) GetMarketsRegionIdOrders(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetMarketsRegionIdOrders", arg0, arg1, arg2)
}

// Mock of HistoryFetcher interface
type MockHistoryFetcher struct {
	ctrl     *gomock.Controller
	recorder *_MockHistoryFetcherRecorder
}

// Recorder for MockHistoryFetcher (not exported)
type _MockHistoryFetcherRecorder struct {
	mock *MockHistoryFetcher
}

func NewMockHistoryFetcher(ctrl *gomock.Controller) *MockHistoryFetcher {
	mock := &MockHistoryFetcher{ctrl: ctrl}
	mock.recorder = &_MockHistoryFetcherRecorder{mock}
	return mock
}

func (_m *MockHistoryFetcher) EXPECT() *_MockHistoryFetcherRecorder {
	return _m.recorder
}

func (_m *MockHistoryFetcher) GetMarketsRegionIdHistory(_param0 int32, _param1 int32, _param2 map[string]interface{}) ([]v1.GetMarketsRegionIdHistory200Ok, *http.Response, error) {
	ret := _m.ctrl.Call(_m, "GetMarketsRegionIdHistory", _param0, _param1, _param2)
	ret0, _ := ret[0].([]v1.GetMarketsRegionIdHistory200Ok)
	ret1, _ := ret[1].(*http.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (_mr *_MockHistoryFetcherRecorder) GetMarketsRegionIdHistory(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetMarketsRegionIdHistory", arg0, arg1, arg2)
}

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

func (_m *MockOrderPublisher) PublishOrder(_param0 *OrderPayload) {
	_m.ctrl.Call(_m, "PublishOrder", _param0)
}

func (_mr *_MockOrderPublisherRecorder) PublishOrder(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "PublishOrder", arg0)
}

func (_m *MockOrderPublisher) PublishStateBegin(_param0 int32) {
	_m.ctrl.Call(_m, "PublishStateBegin", _param0)
}

func (_mr *_MockOrderPublisherRecorder) PublishStateBegin(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "PublishStateBegin", arg0)
}

func (_m *MockOrderPublisher) PublishStateEnd(_param0 int32) {
	_m.ctrl.Call(_m, "PublishStateEnd", _param0)
}

func (_mr *_MockOrderPublisherRecorder) PublishStateEnd(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "PublishStateEnd", arg0)
}