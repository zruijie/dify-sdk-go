package test

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"testing"
	"time"

	"sync"

	"github.com/KevinZhao/dify-sdk-go"
)

var (
	host         = "这里填写你的host"
	apiSecretKey = "这里填写你的api secret key"
)

func TestApi3(t *testing.T) {
	var c = &dify.ClientConfig{
		Host:         host,
		ApiSecretKey: apiSecretKey,
	}
	var client = dify.NewClientWithConfig(c)

	ctx := context.Background()

	var (
		ch  = make(chan dify.ChatMessageStreamChannelResponse)
		err error
	)

	ch, err = client.Api().ChatMessagesStream(ctx, &dify.ChatMessageRequest{
		Query: "你是谁?",
		User:  "这里换成你创建的",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	var (
		strBuilder strings.Builder
		cId        string
	)
	for {
		select {
		case <-ctx.Done():
			t.Log("ctx.Done", strBuilder.String())
			return
		case r, isOpen := <-ch:
			if !isOpen {
				goto done
			}
			strBuilder.WriteString(r.Answer)
			cId = r.ConversationID
			log.Println("Answer2", r.Answer, r.ConversationID, cId, r.ID, r.TaskID)
		}
	}

done:
	t.Log(strBuilder.String())
	t.Log(cId)
}

func TestMessages(t *testing.T) {
	var cId = "ec373942-2d17-4f11-89bb-f9bbf863ebcc"
	var err error
	ctx := context.Background()

	// messages
	var messageReq = &dify.MessagesRequest{
		ConversationID: cId,
		User:           "jiuquan AI",
	}

	var client = dify.NewClient(host, apiSecretKey)

	var msg *dify.MessagesResponse
	if msg, err = client.Api().Messages(ctx, messageReq); err != nil {
		t.Fatal(err.Error())
		return
	}
	j, _ := json.Marshal(msg)
	t.Log(string(j))
}

func TestMessagesFeedbacks(t *testing.T) {
	var client = dify.NewClient(host, apiSecretKey)
	var err error
	ctx := context.Background()

	var id = "72d3dc0f-a6d5-4b5e-8510-bec0611a6048"

	var res *dify.MessagesFeedbacksResponse
	if res, err = client.Api().MessagesFeedbacks(ctx, &dify.MessagesFeedbacksRequest{
		MessageID: id,
		Rating:    dify.FeedbackLike,
		User:      "jiuquan AI",
	}); err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)

	log.Println(string(j))
}

func TestConversations(t *testing.T) {
	var client = dify.NewClient(host, apiSecretKey)
	var err error
	ctx := context.Background()

	var res *dify.ConversationsResponse
	if res, err = client.Api().Conversations(ctx, &dify.ConversationsRequest{
		User: "jiuquan AI",
	}); err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)

	log.Println(string(j))
}

func TestConversationsRename(t *testing.T) {
	var client = dify.NewClient(host, apiSecretKey)
	var err error
	ctx := context.Background()

	var res *dify.ConversationsRenamingResponse
	if res, err = client.Api().ConversationsRenaming(ctx, &dify.ConversationsRenamingRequest{
		ConversationID: "ec373942-2d17-4f11-89bb-f9bbf863ebcc",
		Name:           "rename!!!",
		User:           "jiuquan AI",
	}); err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)

	log.Println(string(j))
}

func TestParameters(t *testing.T) {
	var client = dify.NewClient(host, apiSecretKey)
	var err error
	ctx := context.Background()

	var res *dify.ParametersResponse
	if res, err = client.Api().Parameters(ctx, &dify.ParametersRequest{
		User: "jiuquan AI",
	}); err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)

	log.Println(string(j))
}

func TestRunWorkflow(t *testing.T) {
	client := dify.NewClient(host, apiSecretKey)

	workflowReq := dify.WorkflowRequest{
		Inputs: map[string]interface{}{
			"image_url": "Some image url",
		},
		ResponseMode: "blocking", // 设置为阻塞模式以等待完整的返回
		User:         "Zhaokm@AWS",
	}

	resp, err := client.API().RunWorkflow(context.Background(), workflowReq)

	if err != nil {
		t.Fatalf("RunWorkflow encountered an error: %v", err)
	}

	if resp.WorkflowRunID == "" {
		t.Errorf("Expected non-empty WorkflowRunID, got empty")
	}
	if resp.TaskID == "" {
		t.Errorf("Expected non-empty TaskID, got empty")
	}

	if resp.Data.Status != "succeeded" {
		t.Errorf("Expected workflow status 'completed', got: %v", resp.Data.Status)
	}
	if len(resp.Data.Outputs) == 0 {
		t.Errorf("Expected outputs, but got none")
	}

	if resp.Data.ElapsedTime <= 0 {
		t.Errorf("Expected positive ElapsedTime, but got: %v", resp.Data.ElapsedTime)
	}
	if resp.Data.TotalSteps <= 0 {
		t.Errorf("Expected positive TotalSteps, but got: %v", resp.Data.TotalSteps)
	}

	t.Logf("Received workflow response: %+v", resp)
}

func TestRunWorkflowStreaming(t *testing.T) {
	client := dify.NewClient(host, apiSecretKey)

	workflowReq := dify.WorkflowRequest{
		Inputs: map[string]interface{}{
			"image_url": "https://assets.cnzlerp.com/test/aoolia/1-1.jpg",
		},
		ResponseMode: "streaming",
		User:         "Zhaokm@AWS",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var (
		finalResponse dify.StreamingResponse
		stepCount     int
		hasSucceeded  bool
		mu            sync.Mutex
	)

	handler := func(streamResp dify.StreamingResponse) {
		mu.Lock()
		defer mu.Unlock()

		stepCount++
		t.Logf("Received step %d response: %+v", stepCount, streamResp)

		if streamResp.Data.Status == "succeeded" {
			hasSucceeded = true
		}

		finalResponse = streamResp
	}

	err := client.API().RunStreamWorkflow(ctx, workflowReq, handler)

	// 检查是否有错误
	if err != nil {
		t.Fatalf("RunStreamWorkflow encountered an error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if stepCount == 0 {
		t.Errorf("Expected to receive at least one streaming response, but got none")
	}

	if finalResponse.WorkflowRunID == "" {
		t.Errorf("Expected non-empty WorkflowRunID, got empty")
	}
	if finalResponse.TaskID == "" {
		t.Errorf("Expected non-empty TaskID, got empty")
	}

	if !hasSucceeded {
		t.Errorf("Expected workflow status to be 'succeeded', but got: %v", finalResponse.Data.Status)
	}

	if len(finalResponse.Data.Outputs) == 0 {
		t.Errorf("Expected outputs in the final response, but got none")
	}

	// 检查时间和步骤等其他字段
	if finalResponse.Data.ElapsedTime <= 0 {
		t.Errorf("Expected positive ElapsedTime, but got: %v", finalResponse.Data.ElapsedTime)
	}

	t.Logf("Final streaming workflow response: %+v", finalResponse)
}
