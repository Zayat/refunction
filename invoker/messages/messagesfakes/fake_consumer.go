// Code generated by counterfeiter. DO NOT EDIT.
package messagesfakes

import (
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/ostenbom/refunction/invoker/messages"
)

type FakeConsumer struct {
	ReadMessageStub        func(time.Duration) (*kafka.Message, error)
	readMessageMutex       sync.RWMutex
	readMessageArgsForCall []struct {
		arg1 time.Duration
	}
	readMessageReturns struct {
		result1 *kafka.Message
		result2 error
	}
	readMessageReturnsOnCall map[int]struct {
		result1 *kafka.Message
		result2 error
	}
	SubscribeStub        func(string, kafka.RebalanceCb) error
	subscribeMutex       sync.RWMutex
	subscribeArgsForCall []struct {
		arg1 string
		arg2 kafka.RebalanceCb
	}
	subscribeReturns struct {
		result1 error
	}
	subscribeReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeConsumer) ReadMessage(arg1 time.Duration) (*kafka.Message, error) {
	fake.readMessageMutex.Lock()
	ret, specificReturn := fake.readMessageReturnsOnCall[len(fake.readMessageArgsForCall)]
	fake.readMessageArgsForCall = append(fake.readMessageArgsForCall, struct {
		arg1 time.Duration
	}{arg1})
	fake.recordInvocation("ReadMessage", []interface{}{arg1})
	fake.readMessageMutex.Unlock()
	if fake.ReadMessageStub != nil {
		return fake.ReadMessageStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.readMessageReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeConsumer) ReadMessageCallCount() int {
	fake.readMessageMutex.RLock()
	defer fake.readMessageMutex.RUnlock()
	return len(fake.readMessageArgsForCall)
}

func (fake *FakeConsumer) ReadMessageCalls(stub func(time.Duration) (*kafka.Message, error)) {
	fake.readMessageMutex.Lock()
	defer fake.readMessageMutex.Unlock()
	fake.ReadMessageStub = stub
}

func (fake *FakeConsumer) ReadMessageArgsForCall(i int) time.Duration {
	fake.readMessageMutex.RLock()
	defer fake.readMessageMutex.RUnlock()
	argsForCall := fake.readMessageArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeConsumer) ReadMessageReturns(result1 *kafka.Message, result2 error) {
	fake.readMessageMutex.Lock()
	defer fake.readMessageMutex.Unlock()
	fake.ReadMessageStub = nil
	fake.readMessageReturns = struct {
		result1 *kafka.Message
		result2 error
	}{result1, result2}
}

func (fake *FakeConsumer) ReadMessageReturnsOnCall(i int, result1 *kafka.Message, result2 error) {
	fake.readMessageMutex.Lock()
	defer fake.readMessageMutex.Unlock()
	fake.ReadMessageStub = nil
	if fake.readMessageReturnsOnCall == nil {
		fake.readMessageReturnsOnCall = make(map[int]struct {
			result1 *kafka.Message
			result2 error
		})
	}
	fake.readMessageReturnsOnCall[i] = struct {
		result1 *kafka.Message
		result2 error
	}{result1, result2}
}

func (fake *FakeConsumer) Subscribe(arg1 string, arg2 kafka.RebalanceCb) error {
	fake.subscribeMutex.Lock()
	ret, specificReturn := fake.subscribeReturnsOnCall[len(fake.subscribeArgsForCall)]
	fake.subscribeArgsForCall = append(fake.subscribeArgsForCall, struct {
		arg1 string
		arg2 kafka.RebalanceCb
	}{arg1, arg2})
	fake.recordInvocation("Subscribe", []interface{}{arg1, arg2})
	fake.subscribeMutex.Unlock()
	if fake.SubscribeStub != nil {
		return fake.SubscribeStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.subscribeReturns
	return fakeReturns.result1
}

func (fake *FakeConsumer) SubscribeCallCount() int {
	fake.subscribeMutex.RLock()
	defer fake.subscribeMutex.RUnlock()
	return len(fake.subscribeArgsForCall)
}

func (fake *FakeConsumer) SubscribeCalls(stub func(string, kafka.RebalanceCb) error) {
	fake.subscribeMutex.Lock()
	defer fake.subscribeMutex.Unlock()
	fake.SubscribeStub = stub
}

func (fake *FakeConsumer) SubscribeArgsForCall(i int) (string, kafka.RebalanceCb) {
	fake.subscribeMutex.RLock()
	defer fake.subscribeMutex.RUnlock()
	argsForCall := fake.subscribeArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeConsumer) SubscribeReturns(result1 error) {
	fake.subscribeMutex.Lock()
	defer fake.subscribeMutex.Unlock()
	fake.SubscribeStub = nil
	fake.subscribeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeConsumer) SubscribeReturnsOnCall(i int, result1 error) {
	fake.subscribeMutex.Lock()
	defer fake.subscribeMutex.Unlock()
	fake.SubscribeStub = nil
	if fake.subscribeReturnsOnCall == nil {
		fake.subscribeReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.subscribeReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeConsumer) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.readMessageMutex.RLock()
	defer fake.readMessageMutex.RUnlock()
	fake.subscribeMutex.RLock()
	defer fake.subscribeMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeConsumer) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ messages.Consumer = new(FakeConsumer)