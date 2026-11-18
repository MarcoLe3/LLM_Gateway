package job

type Message struct {
	Role 	string `json:"role"`
	Content string `json:"content"`
}

type RequestPayload struct {
	Model		string 		`json:"model"`
	Messages	[]Message	`json:"messages"`
	MaxTokens	int			`json:"max_tokens, omitempty"`
	Tempature 	*float64
}

type ResponsePayload struct {
	Text	string `json:"text"`
	
}