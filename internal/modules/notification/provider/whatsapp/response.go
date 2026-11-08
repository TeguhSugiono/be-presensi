package whatsapp

import "encoding/json"

type SendMessageResponse struct {
	Status    bool            `json:"status"`
	Detail    string          `json:"detail"`
	Reason    string          `json:"reason"`
	Process   string          `json:"process"`
	RequestID int64           `json:"requestid"`
	ID        json.RawMessage `json:"id"`
	Target    json.RawMessage `json:"target"`
}
