package realtime

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type Client struct {
	userID string
	send   chan Event
}
