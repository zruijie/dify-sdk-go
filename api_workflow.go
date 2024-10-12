package dify

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DifyClient 用于封装对 Dify API 的调用
type DifyClient struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

// NewDifyClient 创建一个新的 Dify API 客户端
func NewDifyClient(apiKey, baseURL string) *DifyClient {
	return &DifyClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Client:  &http.Client{},
	}
}

// WorkflowRequest 表示发起 Workflow 的请求体
type WorkflowRequest struct {
	Inputs       map[string]interface{} `json:"inputs"`
	ResponseMode string                 `json:"response_mode"`
	User         string                 `json:"user"`
}

// WorkflowResponse 表示 Workflow 返回的数据结构（阻塞模式）
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

// StreamingResponse 表示流式返回的数据结构
type StreamingResponse struct {
	Event          string `json:"event"`
	TaskID         string `json:"task_id"`
	WorkflowRunID  string `json:"workflow_run_id"`
	SequenceNumber int    `json:"sequence_number"`
	Data           struct {
		ID          string  `json:"id"`
		WorkflowID  string  `json:"workflow_id"`
		NodeID      string  `json:"node_id"`
		NodeType    string  `json:"node_type"`
		Title       string  `json:"title"`
		Index       int     `json:"index"`
		Predecessor string  `json:"predecessor_node_id,omitempty"`
		Outputs     string  `json:"outputs,omitempty"`
		Status      string  `json:"status"`
		ElapsedTime float64 `json:"elapsed_time"`
		TotalTokens int     `json:"total_tokens,omitempty"`
		TotalPrice  float64 `json:"total_price,omitempty"`
		Currency    string  `json:"currency,omitempty"`
		CreatedAt   int64   `json:"created_at"`
	} `json:"data"`
}

// RunWorkflow 执行指定的 Workflow（阻塞模式）
func (c *DifyClient) RunWorkflow(request WorkflowRequest) (*WorkflowResponse, error) {
	url := fmt.Sprintf("%s/workflows/run", c.BaseURL)

	// 创建请求体
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	// 设置头信息
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %s: %s", resp.Status, string(bodyBytes))
	}

	var workflowResp WorkflowResponse
	if err := json.NewDecoder(resp.Body).Decode(&workflowResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workflowResp, nil
}

// StreamWorkflow 执行指定的 Workflow（流式模式）
func (c *DifyClient) StreamWorkflow(request WorkflowRequest, handler func(StreamingResponse)) error {
	url := fmt.Sprintf("%s/workflows/run", c.BaseURL)

	// 创建请求体
	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create new request: %w", err)
	}

	// 设置头信息
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %s: %s", resp.Status, string(bodyBytes))
	}

	// 读取和处理流式响应
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading streaming response: %w", err)
		}

		// 解析流式数据
		if len(line) > 6 && string(line[:6]) == "data: " {
			var streamResp StreamingResponse
			err := json.Unmarshal(line[6:], &streamResp)
			if err != nil {
				fmt.Println("Error decoding streaming response:", err)
				continue
			}

			// 调用处理函数处理每个流式响应块
			handler(streamResp)
		}
	}

	return nil
}
