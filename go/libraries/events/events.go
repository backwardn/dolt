package events

import (
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"

	eventsapi "github.com/liquidata-inc/dolt/go/gen/proto/dolt/services/eventsapi_v1alpha1"
)

// EventNowFunc function is used to get the current time and can be overridden for testing.
var EventNowFunc = time.Now

func nowTimestamp() *timestamp.Timestamp {
	now := EventNowFunc()
	nanos := int32(now.UnixNano() % int64(time.Second))

	return &timestamp.Timestamp{Seconds: now.Unix(), Nanos: nanos}
}

// Event is an event to be added to a collector and logged
type Event struct {
	ce         *eventsapi.ClientEvent
	m          *sync.Mutex
	attributes map[eventsapi.AttributeID]string
}

// NewEvent creates an Event of a given type.  The event creation time is recorded as the start time for the event.
// When the event is passed to a collector's CloseEventAndAdd method the end time of the event is recorded
func NewEvent(ceType eventsapi.ClientEventType) *Event {
	return &Event{
		ce: &eventsapi.ClientEvent{
			Id:        uuid.New().String(),
			StartTime: nowTimestamp(),
			Type:      ceType,
		},
		m:          &sync.Mutex{},
		attributes: make(map[eventsapi.AttributeID]string),
	}
}

// AddMetric adds a metric to the event.  This method is thread safe.
func (evt *Event) AddMetric(em EventMetric) {
	evt.m.Lock()
	defer evt.m.Unlock()

	evt.ce.Metrics = append(evt.ce.Metrics, em.AsClientEventMetric())
}

// SetAttribute adds an attribute to the event.  This method is thread safe
func (evt *Event) SetAttribute(attID eventsapi.AttributeID, attVal string) {
	evt.m.Lock()
	defer evt.m.Unlock()

	evt.attributes[attID] = attVal
}

// GetAttribute adds an attribute to the event. This method is thread safe
func (evt *Event) GetAttribute(attID eventsapi.AttributeID) string {
	evt.m.Lock()
	defer evt.m.Unlock()

	if val, ok := evt.attributes[attID]; ok {
		return val
	}

	return ""
}

// GetClientEventType returns a pointer to the Client Event. This method is thread safe
func (evt *Event) GetClientEventType() eventsapi.ClientEventType {
	evt.m.Lock()
	defer evt.m.Unlock()

	return evt.ce.Type
}

func (evt *Event) close() *eventsapi.ClientEvent {
	if evt.ce == nil {
		panic("multiple close calls for the same event.")
	}

	evt.m.Lock()
	defer evt.m.Unlock()

	evt.ce.EndTime = nowTimestamp()

	for k, v := range evt.attributes {
		evt.ce.Attributes = append(evt.ce.Attributes, &eventsapi.ClientEventAttribute{Id: k, Value: v})
	}

	ce := evt.ce
	evt.ce = nil

	return ce
}
