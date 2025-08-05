package model

type WsMessage struct {
	Type    string
	Payload string
}

type ScaleDataRequest struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
	Type  string  `json:"type"`
}
