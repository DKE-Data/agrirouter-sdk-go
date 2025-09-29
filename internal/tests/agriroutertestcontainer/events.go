// Package agriroutertestcontainer is a testcontainers module for the agrirouter service.
package agriroutertestcontainer

import (
	"fmt"
	"log"
	"sync"

	"github.com/stretchr/testify/assert"
)

// TestedEndpointPutData represents the data structure for an event
// that is expected when an endpoint is put in the agrirouter test container.
type TestedEndpointPutData struct {
	ExternalID string `json:"externalId"`
}

const (
	// PutEndpointTestEvent happens when an endpoint is put in the test container.
	PutEndpointTestEvent = "putEndpoint"

	// SendMessagesTestEvent happens when messages are sent in the test container.
	SendMessagesTestEvent = "sendMessages"
)

// TestEvent represents a single event happened in test container.
type TestEvent struct {
	Data      string
	EventType string
}

// TestEvents helps to test whether certain events have happened during tests.
type TestEvents struct {
	list []TestEvent

	expectationIndex int
	expectations     []func(t assert.TestingT) error

	mu sync.Mutex
}

func (e *TestEvents) add(evType string, data string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.list = append(e.list, TestEvent{
		Data:      data,
		EventType: evType,
	})
}

// ErrExpectationFailed is an error that indicates that an expected event was not received.
var ErrExpectationFailed = fmt.Errorf("expectation failed")

// Expect adds an expectation for an event of type evType with the given data.
func (e *TestEvents) Expect(
	evType string,
	data string,
) {
	index := e.expectationIndex
	e.expectationIndex++

	e.expectations = append(e.expectations, func(t assert.TestingT) error {
		log.Printf("Checking expectation %d: event type %s with data %s, have %d events", index, evType, data, len(e.list))
		if index >= len(e.list) {
			return fmt.Errorf(
				"%w: expected event %s with data %s, but no more events are available",
				ErrExpectationFailed,
				evType,
				data,
			)
		}
		if !assert.JSONEq(t, data, e.list[index].Data) {
			return fmt.Errorf(
				"%w: expected event data %s, got %s",
				ErrExpectationFailed,
				data,
				e.list[index].Data,
			)
		}

		return nil
	})
}

// CheckExpectations checks if all expected events were received and resets the expectations.
func (e *TestEvents) CheckExpectations(t assert.TestingT) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, expectation := range e.expectations {
		if err := expectation(t); err != nil {
			return err
		}
	}
	e.list = nil
	e.expectations = nil
	e.expectationIndex = 0
	return nil
}
