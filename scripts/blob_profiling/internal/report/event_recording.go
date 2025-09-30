package report

import (
	"runtime"
	"time"
)

// RecordEvent records an event with context
func (pr *ProfilerReport) RecordEvent(eventType EventType, component, message string, err error, context map[string]interface{}) {
	event := Event{
		Timestamp: time.Now(),
		EventType: eventType,
		Component: component,
		Message:   message,
		Context:   context,
	}

	if err != nil {
		event.Error = err.Error()
		event.StackTrace = stackTrace()
	}

	pr.data.Events = append(pr.data.Events, event)
}

// EventCount returns the total number of events recorded
func (pr *ProfilerReport) EventCount() int {
	return len(pr.data.Events)
}

// ErrorCount returns the number of error events
func (pr *ProfilerReport) ErrorCount() int {
	count := 0
	for _, event := range pr.data.Events {
		switch event.EventType {
		case EventTypeError, EventTypeCasting, EventTypeCatching, EventTypeConnection, EventTypeParquet, EventTypeSequencer, EventTypeBlobpool:
			count++
		}
	}
	return count
}

// Helper function for stack traces
func stackTrace() string {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, 2*len(buf))
	}
}
