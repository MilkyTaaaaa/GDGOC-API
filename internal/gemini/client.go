package gemini

import (
    "context"
    "fmt"
    "log"

    "github.com/google/generative-ai-go/genai"  // ✅ PAKAI INI
    "google.golang.org/api/option"              // ✅ PAKAI INI
)

type Client struct {
    client *genai.Client
    ctx    context.Context
}

func NewClient(apiKey string) (*Client, error) {
    if apiKey == "" {
        return nil, fmt.Errorf("Gemini API key is required")
    }

    ctx := context.Background()
    
    // ✅ GUNAKAN PACKAGE YANG BENAR
    client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
    if err != nil {
        return nil, fmt.Errorf("failed to create Gemini client: %v", err)
    }
    
    log.Println("✅ Gemini client initialized successfully")
    return &Client{
        client: client,
        ctx:    ctx,
    }, nil
}

//tutup gemini

func (c *Client) Close() error {
    if c.client != nil {
        return c.client.Close()
    }
    return nil
}