package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

const fonnteSendURL = "https://api.fonnte.com/send"

// ProviderRejectedError means Fonnte explicitly rejected the message.
type ProviderRejectedError struct {
	Reason string
}

func (e *ProviderRejectedError) Error() string {
	return fmt.Sprintf("fonnte rejected message: %s", e.Reason)
}

// DeliveryUncertainError means the request may have been accepted,
// but the client cannot confirm the delivery result.
type DeliveryUncertainError struct {
	Err error
}

func (e *DeliveryUncertainError) Error() string {
	return fmt.Sprintf("fonnte delivery status is uncertain: %v", e.Err)
}

func (e *DeliveryUncertainError) Unwrap() error {
	return e.Err
}

// IsProviderRejected checks whether Fonnte explicitly rejected the message.
func IsProviderRejected(err error) bool {
	var rejectedErr *ProviderRejectedError
	return errors.As(err, &rejectedErr)
}

type WhatsAppClient struct {
	httpClient *http.Client
}

func NewWhatsAppClient() *WhatsAppClient {
	return &WhatsAppClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *WhatsAppClient) SendMessage(
	ctx context.Context,
	token string,
	req *SendMessageRequest,
) (*SendMessageResponse, error) {
	if token == "" {
		return nil, fmt.Errorf("fonnte token is required")
	}

	if req == nil || req.Target == "" || req.Message == "" {
		return nil, fmt.Errorf("target and message are required")
	}

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("target", req.Target); err != nil {
		return nil, fmt.Errorf("failed to write target: %w", err)
	}

	if err := writer.WriteField("message", req.Message); err != nil {
		return nil, fmt.Errorf("failed to write message: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fonnteSendURL,
		&body,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create fonnte request: %w", err)
	}

	httpReq.Header.Set("Authorization", token)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &DeliveryUncertainError{
			Err: fmt.Errorf("failed to connect to fonnte: %w", err),
		}
	}
	defer httpResp.Body.Close()

	responseBody, err := io.ReadAll(
		io.LimitReader(httpResp.Body, 1<<20),
	)
	if err != nil {
		return nil, &DeliveryUncertainError{
			Err: fmt.Errorf("failed to read fonnte response: %w", err),
		}
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, &DeliveryUncertainError{
			Err: fmt.Errorf("fonnte returned HTTP status %d", httpResp.StatusCode),
		}
	}

	var result SendMessageResponse

	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, &DeliveryUncertainError{
			Err: fmt.Errorf("invalid fonnte response: %w", err),
		}
	}

	if !result.Status {
		return &result, &ProviderRejectedError{
			Reason: result.Reason,
		}
	}

	return &result, nil
}
