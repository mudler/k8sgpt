package ai

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/viper"

	"github.com/sashabaranov/go-openai"
)

const (
	default_prompt = "Simplify the following Kubernetes error message and provide a solution in %s: %s"
	prompt_a       = "Read the following input %s and provide possible scenarios for remediation in %s"
	prompt_b       = "Considering the following input from the Kubernetes resource %s and the error message %s, provide possible scenarios for remediation in %s"
	prompt_c       = "Reading the following %s error message and it's accompanying log message %s, how would you simplify this message?"
)

type OpenAIClient struct {
	client   *openai.Client
	language string
	model    string
}

func (c *OpenAIClient) Configure(token string, model string, language string) error {
	client := openai.NewClient(token)
	if client == nil {
		return errors.New("error creating OpenAI client")
	}
	c.language = language
	c.client = client
	c.model = model
	return nil
}

func (c *OpenAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	// Create a completion request
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: fmt.Sprintf(default_prompt, c.language, prompt),
			},
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func (a *OpenAIClient) Parse(ctx context.Context, prompt []string, nocache bool) (string, error) {
	inputKey := strings.Join(prompt, " ")
	// Check for cached data
	sEnc := base64.StdEncoding.EncodeToString([]byte(inputKey))
	cacheKey := util.GetCacheKey(a.GetName(), sEnc)
	// find in viper cache
	if viper.IsSet(cacheKey) && !nocache {
		// retrieve data from cache
		response := viper.GetString(cacheKey)
		if response == "" {
			color.Red("error retrieving cached data")
			return "", nil
		}
		output, err := base64.StdEncoding.DecodeString(response)
		if err != nil {
			color.Red("error decoding cached data: %v", err)
			return "", nil
		}
		return string(output), nil
	}

	response, err := a.GetCompletion(ctx, inputKey)
	if err != nil {
		return "", err
	}

	if !viper.IsSet(cacheKey) || nocache {
		viper.Set(cacheKey, base64.StdEncoding.EncodeToString([]byte(response)))
		if err := viper.WriteConfig(); err != nil {
			color.Red("error writing config: %v", err)
			return "", nil
		}
	}
	return response, nil
}

func (a *OpenAIClient) GetName() string {
	return "openai"
}
