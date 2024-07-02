package generate

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/sashabaranov/go-openai"
	"google.golang.org/api/option"
)

type Platform string

const (
	PlatformOpenAI = Platform("openai")
	PlatformGemini = Platform("gemini")
)

type Client struct {
	platform Platform
	apiKey   string
	baseURL  string
	proxy    string
	// 模型名
	model string
}

func New(platform Platform, apiKey string, baseURL string, proxy string, model string) *Client {
	return &Client{
		platform: platform,
		apiKey:   apiKey,
		baseURL:  baseURL,
		proxy:    proxy,
		model:    model,
	}
}

func (client *Client) SendMessage(prompt string) (string, error) {
	switch client.platform {
	case PlatformOpenAI:
		return client.sendOpenAIMessage(prompt)
	case PlatformGemini:
		return client.sendGeminiMessage(prompt)
	default:
		return "", errors.New("unknown client type")
	}
}

// Gemini 消息
func (client *Client) sendGeminiMessage(prompt string) (string, error) {
	httpClient := &http.Client{
		Transport: &ProxyRoundTripper{
			APIKey:   client.apiKey,
			ProxyURL: client.proxy,
		},
		Timeout: 30 * time.Second,
	}

	ctx := context.Background()
	options := []option.ClientOption{}
	options = append(options, option.WithHTTPClient(httpClient))
	options = append(options, option.WithAPIKey(client.apiKey))

	if len(client.baseURL) > 0 {
		options = append(options, option.WithEndpoint(client.baseURL))
	}

	genaiClient, err := genai.NewClient(ctx, options...)
	if err != nil {
		return "", err
	}
	defer genaiClient.Close()

	// 设置模型
	modelName := client.model
	if len(client.model) == 0 {
		modelName = "gemini-1.5-flash"
	}

	model := genaiClient.GenerativeModel(modelName)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 {
		return "", errors.New("no candidates found")
	}

	content := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			content += string(text)
		}
	}

	return content, nil
}

// openai 消息
func (client *Client) sendOpenAIMessage(prompt string) (string, error) {
	config := openai.DefaultConfig(client.apiKey)

	if len(client.baseURL) > 0 {
		config.BaseURL = client.baseURL
		config.APIType = "CUSTOM"
	}

	if len(client.proxy) > 0 {
		proxyURL, err := url.Parse(client.proxy)
		if err != nil {
			return "", err
		}
		transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		// 设置 30s 超时
		config.HTTPClient = &http.Client{Transport: transport, Timeout: time.Second * 30}
	}

	openaiClient := openai.NewClientWithConfig(config)

	modelName := client.model
	if len(client.model) == 0 {
		modelName = openai.GPT3Dot5Turbo
	}
	resp, err := openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       modelName,
			Temperature: 0.8,
			Messages:    []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: prompt}},
		})

	if err != nil {
		return "", err
	}

	message := resp.Choices[0].Message.Content
	return message, nil
}

// ProxyRoundTripper is an implementation of http.RoundTripper that supports
// setting a proxy server URL for genai clients. This type should be used with
// a custom http.Client that's passed to WithHTTPClient. For such clients,
// WithAPIKey doesn't apply so the key has to be explicitly set here.
type ProxyRoundTripper struct {
	APIKey   string
	ProxyURL string
}

func (t *ProxyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if t.ProxyURL != "" {
		proxyURL, err := url.Parse(t.ProxyURL)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	newReq := req.Clone(req.Context())
	vals := newReq.URL.Query()
	vals.Set("key", t.APIKey)
	newReq.URL.RawQuery = vals.Encode()

	newReq.Header.Set("Content-Type", "application/json")
	newReq.Header.Set("User-Agent", "google-api-go-client")

	resp, err := transport.RoundTrip(newReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
