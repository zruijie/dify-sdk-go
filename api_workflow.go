package dify

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// WorkflowRequest
type WorkflowRequest struct {
	Inputs       map[string]interface{} `json:"inputs"`
	ResponseMode string                 `json:"response_mode"`
	User         string                 `json:"user"`
}

// WorkflowResponse
type WorkflowResponse struct {
	WorkflowRunID string `json:"workflow_run_id"`
	TaskID        string `json:"task_id"`
	Data          struct {
		ID          string                 `json:"id"`
		WorkflowID  string                 `json:"workflow_id"`
		Status      string                 `json:"status"`
		Outputs     map[string]interface{} `json:"outputs"`
		Error       *string                `json:"error,omitempty"`
		ElapsedTime float64                `json:"elapsed_time"`
		TotalTokens int                    `json:"total_tokens"`
		TotalSteps  int                    `json:"total_steps"`
		CreatedAt   int64                  `json:"created_at"`
		FinishedAt  int64                  `json:"finished_at"`
	} `json:"data"`
}

// StreamingResponse
type StreamingResponse struct {
	Event          string `json:"event"`
	TaskID         string `json:"task_id"`
	WorkflowRunID  string `json:"workflow_run_id"`
	SequenceNumber int    `json:"sequence_number"`
	Data           struct {
		ID          string                 `json:"id"`
		WorkflowID  string                 `json:"workflow_id"`
		NodeID      string                 `json:"node_id"`
		NodeType    string                 `json:"node_type"`
		Title       string                 `json:"title"`
		Index       int                    `json:"index"`
		Predecessor string                 `json:"predecessor_node_id,omitempty"`
		Outputs     map[string]interface{} `json:"outputs,omitempty"`
		Status      string                 `json:"status"`
		ElapsedTime float64                `json:"elapsed_time"`
		TotalTokens int                    `json:"total_tokens,omitempty"`
		TotalPrice  float64                `json:"total_price,omitempty"`
		Currency    string                 `json:"currency,omitempty"`
		CreatedAt   int64                  `json:"created_at"`
	} `json:"data"`
}

// RunWorkflow
func (api *API) RunWorkflow(ctx context.Context, request WorkflowRequest) (*WorkflowResponse, error) {
	req, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/workflows/run", request)
	if err != nil {
		return nil, fmt.Errorf("failed to create base request: %w", err)
	}

	fmt.Print(req)

	resp, err := api.c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %s: %s", resp.Status, readResponseBody(resp.Body))
	}

	var workflowResp WorkflowResponse
	if err := json.NewDecoder(resp.Body).Decode(&workflowResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workflowResp, nil
}

// RunStreamWorkflow
func (api *API) RunStreamWorkflow(ctx context.Context, request WorkflowRequest, handler func(StreamingResponse)) error {
	// 直接传递 request 对象，而不是序列化后的字节流
	req, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/workflows/run", request)
	if err != nil {
		return err
	}

	// 记录请求信息
	fmt.Printf("Request: %v\n", req)

	resp, err := api.c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status %s: %s", resp.Status, readResponseBody(resp.Body))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading streaming response: %w", err)
		}

		if len(line) > 6 && string(line[:6]) == "data: " {
			var streamResp StreamingResponse
			if err := json.Unmarshal(line[6:], &streamResp); err != nil {
				fmt.Println("Error decoding streaming response:", err)
				continue
			}

			handler(streamResp)
		}
	}

	return nil
}

// Helper function to read response body
func readResponseBody(body io.Reader) string {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return "unable to read body"
	}
	return string(bodyBytes)
}
