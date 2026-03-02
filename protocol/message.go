package protocol

type ProduceRequest struct {
    Type      string `json:"type"`
    Topic     string `json:"topic"`
    Partition int    `json:"partition"`
    Message   string `json:"message"`
}

type FetchRequest struct {
    Type      string `json:"type"`
    Topic     string `json:"topic"`
    Partition int    `json:"partition"`
    Offset    int64  `json:"offset"`
}