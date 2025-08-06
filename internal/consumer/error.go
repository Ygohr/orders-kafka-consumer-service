package consumer

type ProcessingError struct {
	Message   string `json:"message"`
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
	Error     string `json:"error"`
}
