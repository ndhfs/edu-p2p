package dto

type Common struct {
	Message *Message `json:"message,omitempty"`
}

func NewTextMessage(name string, text string) Common {
	return Common{
		Message: &Message{
			From: Client{
				Name: name,
			},
			Text: text,
		},
	}
}

type Message struct {
	From Client `json:"from,omitempty"`
	Text string `json:"text,omitempty"`
}

type Client struct {
	Id   uint   `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
