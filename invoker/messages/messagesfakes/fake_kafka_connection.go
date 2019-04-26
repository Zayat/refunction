// Code generated by counterfeiter. DO NOT EDIT.
package messagesfakes

import (
	"sync"

	"github.com/ostenbom/refunction/invoker/messages"
	kafka "github.com/segmentio/kafka-go"
)

type FakeKafkaConnection struct {
	CloseStub        func() error
	closeMutex       sync.RWMutex
	closeArgsForCall []struct {
	}
	closeReturns struct {
		result1 error
	}
	closeReturnsOnCall map[int]struct {
		result1 error
	}
	CreateTopicsStub        func(...kafka.TopicConfig) error
	createTopicsMutex       sync.RWMutex
	createTopicsArgsForCall []struct {
		arg1 []kafka.TopicConfig
	}
	createTopicsReturns struct {
		result1 error
	}
	createTopicsReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeKafkaConnection) Close() error {
	fake.closeMutex.Lock()
	ret, specificReturn := fake.closeReturnsOnCall[len(fake.closeArgsForCall)]
	fake.closeArgsForCall = append(fake.closeArgsForCall, struct {
	}{})
	fake.recordInvocation("Close", []interface{}{})
	fake.closeMutex.Unlock()
	if fake.CloseStub != nil {
		return fake.CloseStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.closeReturns
	return fakeReturns.result1
}

func (fake *FakeKafkaConnection) CloseCallCount() int {
	fake.closeMutex.RLock()
	defer fake.closeMutex.RUnlock()
	return len(fake.closeArgsForCall)
}

func (fake *FakeKafkaConnection) CloseCalls(stub func() error) {
	fake.closeMutex.Lock()
	defer fake.closeMutex.Unlock()
	fake.CloseStub = stub
}

func (fake *FakeKafkaConnection) CloseReturns(result1 error) {
	fake.closeMutex.Lock()
	defer fake.closeMutex.Unlock()
	fake.CloseStub = nil
	fake.closeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeKafkaConnection) CloseReturnsOnCall(i int, result1 error) {
	fake.closeMutex.Lock()
	defer fake.closeMutex.Unlock()
	fake.CloseStub = nil
	if fake.closeReturnsOnCall == nil {
		fake.closeReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.closeReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeKafkaConnection) CreateTopics(arg1 ...kafka.TopicConfig) error {
	fake.createTopicsMutex.Lock()
	ret, specificReturn := fake.createTopicsReturnsOnCall[len(fake.createTopicsArgsForCall)]
	fake.createTopicsArgsForCall = append(fake.createTopicsArgsForCall, struct {
		arg1 []kafka.TopicConfig
	}{arg1})
	fake.recordInvocation("CreateTopics", []interface{}{arg1})
	fake.createTopicsMutex.Unlock()
	if fake.CreateTopicsStub != nil {
		return fake.CreateTopicsStub(arg1...)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.createTopicsReturns
	return fakeReturns.result1
}

func (fake *FakeKafkaConnection) CreateTopicsCallCount() int {
	fake.createTopicsMutex.RLock()
	defer fake.createTopicsMutex.RUnlock()
	return len(fake.createTopicsArgsForCall)
}

func (fake *FakeKafkaConnection) CreateTopicsCalls(stub func(...kafka.TopicConfig) error) {
	fake.createTopicsMutex.Lock()
	defer fake.createTopicsMutex.Unlock()
	fake.CreateTopicsStub = stub
}

func (fake *FakeKafkaConnection) CreateTopicsArgsForCall(i int) []kafka.TopicConfig {
	fake.createTopicsMutex.RLock()
	defer fake.createTopicsMutex.RUnlock()
	argsForCall := fake.createTopicsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeKafkaConnection) CreateTopicsReturns(result1 error) {
	fake.createTopicsMutex.Lock()
	defer fake.createTopicsMutex.Unlock()
	fake.CreateTopicsStub = nil
	fake.createTopicsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeKafkaConnection) CreateTopicsReturnsOnCall(i int, result1 error) {
	fake.createTopicsMutex.Lock()
	defer fake.createTopicsMutex.Unlock()
	fake.CreateTopicsStub = nil
	if fake.createTopicsReturnsOnCall == nil {
		fake.createTopicsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.createTopicsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeKafkaConnection) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.closeMutex.RLock()
	defer fake.closeMutex.RUnlock()
	fake.createTopicsMutex.RLock()
	defer fake.createTopicsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeKafkaConnection) recordInvocation(key string, args []interface{}) {
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

var _ messages.KafkaConnection = new(FakeKafkaConnection)
