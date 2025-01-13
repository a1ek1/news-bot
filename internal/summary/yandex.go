package summary

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/sheeiavellie/go-yandexgpt"
)

type YandexSummarizer struct {
	client  *yandexgpt.YandexGPTClient
	prompt  string
	enabled bool
	mu      sync.Mutex
}

func NewYandexSummarizer(apiKey, prompt string) *YandexSummarizer {
	s := &YandexSummarizer{
		client: yandexgpt.NewYandexGPTClientWithAPIKey(apiKey),
		prompt: prompt,
	}

	log.Printf("yandex summarizer is enabled: %v", apiKey != "")

	if apiKey != "" {
		s.enabled = true
	}

	return s
}

func (s *YandexSummarizer) Summarize(text string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled {
		return "", fmt.Errorf("yandex summarizer is disabled")
	}

	request := yandexgpt.YandexGPTRequest{
		ModelURI: yandexgpt.MakeModelURI("b1g64459l652jjcptras", yandexgpt.YandexGPTModelLite),
		CompletionOptions: yandexgpt.YandexGPTCompletionOptions{
			Stream:      false,
			Temperature: 0.7,
			MaxTokens:   2000,
		},
		Messages: []yandexgpt.YandexGPTMessage{
			{
				Role: yandexgpt.YandexGPTMessageRoleSystem,
				Text: s.prompt,
			},
			{
				Role: yandexgpt.YandexGPTMessageRoleUser,
				Text: text,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	resp, err := s.client.GetCompletion(ctx, request)
	if err != nil {
		return "", err
	}

	if len(resp.Result.Alternatives) == 0 {
		return "", errors.New("no choices in yandex response")
	}

	rawSummary := strings.TrimSpace(resp.Result.Alternatives[0].Message.Text)
	if strings.HasSuffix(rawSummary, ".") {
		return rawSummary, nil
	}

	sentences := strings.Split(rawSummary, ".")

	return strings.Join(sentences[:len(sentences)-1], ".") + ".", nil
}
