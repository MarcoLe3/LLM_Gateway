package job

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestPayload struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
	Tempature *float64  `json:"tempature,omitempty"`
}

type ResponsePayload struct {
	Text         string `json:"text"`
	Provider     string `json:"provider"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

const (
	maxMessages     = 100
	maxContentBytes = 100_000
	maxOutputTOkens = 8192
)

var validRoles = map[string]bool{
	"system":    true,
	"user":      true,
	"assistant": true,
}

func (p RequestPayload) Validate() error {
	if len(p.Messages) <= 0 {
		return errors.New("messages must not be empty")
	}

	if len(p.Messages) > maxMessages {
		return fmt.Errorf("too many messages (max %d)", maxMessages)
	}

	totalBytes := 0
	for i, m := range p.Messages {
		if !validRoles[m.Role] {
			return fmt.Errorf("message %d: invalid role %q", i, m.Role)
		}

		if m.Content == "" {
			return fmt.Errorf("message %d: invalid content, content must not be empty", i)
		}

		totalBytes += len(m.Content)
	}

	if totalBytes > maxContentBytes {
		return fmt.Errorf("message too large (max %d bytes)", maxContentBytes)
	}

	return nil
}

func (p RequestPayload) Hash() (string, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
