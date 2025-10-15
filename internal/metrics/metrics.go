package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all the application metrics
type Metrics struct {
	// Room metrics
	RoomsCreatedTotal     prometheus.Counter
	RoomsActive           prometheus.Gauge
	RoomParticipants      prometheus.GaugeVec
	
	// User metrics
	UsersRegisteredTotal  prometheus.Counter
	UsersOnline           prometheus.Gauge
	
	// WebSocket metrics
	WebSocketConnections  prometheus.Gauge
	WebSocketMessagesSent *prometheus.CounterVec
	WebSocketErrorsTotal  prometheus.Counter
	
	// Call metrics
	CallsStartedTotal     prometheus.Counter
	CallsActive           prometheus.Gauge
	CallDurationSeconds   prometheus.Histogram
	
	// Recording metrics
	RecordingsStartedTotal prometheus.Counter
	RecordingsCompletedTotal prometheus.Counter
	RecordingErrorsTotal   prometheus.Counter
	
	// Chat metrics
	ChatMessagesSentTotal prometheus.Counter
}

// AppMetrics is the global metrics instance
var AppMetrics *Metrics

func init() {
	AppMetrics = &Metrics{
		// Room metrics
		RoomsCreatedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "video_call_rooms_created_total",
			Help: "Total number of rooms created",
		}),
		RoomsActive: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "video_call_rooms_active",
			Help: "Number of active rooms",
		}),
		RoomParticipants: *promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "video_call_room_participants",
			Help: "Number of participants in rooms",
		}, []string{"room_id"}),
		
		// User metrics
		UsersRegisteredTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "video_call_users_registered_total",
			Help: "Total number of registered users",
		}),
		UsersOnline: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "video_call_users_online",
			Help: "Number of online users",
		}),
		
		// WebSocket metrics
		WebSocketConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "video_call_websocket_connections",
			Help: "Number of active WebSocket connections",
		}),
		WebSocketMessagesSent: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "video_call_websocket_messages_sent_total",
			Help: "Total number of WebSocket messages sent",
		}, []string{"message_type"}),
		WebSocketErrorsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "video_call_websocket_errors_total",
			Help: "Total number of WebSocket errors",
		}),
		
		// Call metrics
		CallsStartedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "video_call_calls_started_total",
			Help: "Total number of calls started",
		}),
		CallsActive: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "video_call_calls_active",
			Help: "Number of active calls",
		}),
		CallDurationSeconds: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "video_call_call_duration_seconds",
			Help:    "Call duration in seconds",
			Buckets: prometheus.ExponentialBuckets(10, 2, 10), // 10s, 20s, 40s, ..., 5120s
		}),
		
		// Recording metrics
		RecordingsStartedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "video_call_recordings_started_total",
			Help: "Total number of recordings started",
		}),
		RecordingsCompletedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "video_call_recordings_completed_total",
			Help: "Total number of recordings completed",
		}),
		RecordingErrorsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "video_call_recording_errors_total",
			Help: "Total number of recording errors",
		}),
		
		// Chat metrics
		ChatMessagesSentTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "video_call_chat_messages_sent_total",
			Help: "Total number of chat messages sent",
		}),
	}
}

// IncrementRoomsCreated increments the rooms created counter
func (m *Metrics) IncrementRoomsCreated() {
	m.RoomsCreatedTotal.Inc()
}

// SetRoomsActive sets the number of active rooms
func (m *Metrics) SetRoomsActive(count float64) {
	m.RoomsActive.Set(count)
}

// SetRoomParticipants sets the number of participants in a room
func (m *Metrics) SetRoomParticipants(roomID string, count float64) {
	m.RoomParticipants.WithLabelValues(roomID).Set(count)
}

// IncrementUsersRegistered increments the users registered counter
func (m *Metrics) IncrementUsersRegistered() {
	m.UsersRegisteredTotal.Inc()
}

// SetUsersOnline sets the number of online users
func (m *Metrics) SetUsersOnline(count float64) {
	m.UsersOnline.Set(count)
}

// SetWebSocketConnections sets the number of WebSocket connections
func (m *Metrics) SetWebSocketConnections(count float64) {
	m.WebSocketConnections.Set(count)
}

// IncrementWebSocketMessagesSent increments the WebSocket messages sent counter
func (m *Metrics) IncrementWebSocketMessagesSent(messageType string) {
	m.WebSocketMessagesSent.WithLabelValues(messageType).Inc()
}

// IncrementWebSocketErrors increments the WebSocket errors counter
func (m *Metrics) IncrementWebSocketErrors() {
	m.WebSocketErrorsTotal.Inc()
}

// IncrementCallsStarted increments the calls started counter
func (m *Metrics) IncrementCallsStarted() {
	m.CallsStartedTotal.Inc()
}

// SetCallsActive sets the number of active calls
func (m *Metrics) SetCallsActive(count float64) {
	m.CallsActive.Set(count)
}

// ObserveCallDuration observes a call duration
func (m *Metrics) ObserveCallDuration(duration float64) {
	m.CallDurationSeconds.Observe(duration)
}

// IncrementRecordingsStarted increments the recordings started counter
func (m *Metrics) IncrementRecordingsStarted() {
	m.RecordingsStartedTotal.Inc()
}

// IncrementRecordingsCompleted increments the recordings completed counter
func (m *Metrics) IncrementRecordingsCompleted() {
	m.RecordingsCompletedTotal.Inc()
}

// IncrementRecordingErrors increments the recording errors counter
func (m *Metrics) IncrementRecordingErrors() {
	m.RecordingErrorsTotal.Inc()
}

// IncrementChatMessagesSent increments the chat messages sent counter
func (m *Metrics) IncrementChatMessagesSent() {
	m.ChatMessagesSentTotal.Inc()
}