package tcp

const (
	maxMessageLength = 2048 // max length
	//heartbeat        = 0    // heartbeat package
)

/**
 * receive, send struct
 */
type Message struct {
	Type uint
	Data []byte
}
