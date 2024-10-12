package test

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"testing"

	"github.com/langgenius/dify-sdk-go"
)

var (
	//host         = "这里填写你的host"
	//apiSecretKey = "这里填写你的api secret key"
	host         = "https://dify.zhaokm.org"
	apiSecretKey = "app-h0Xpt3bR74UImZK2sSgbhIoh"
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
	var client = dify.NewClient(host, apiSecretKey)
	var err error
	ctx := context.Background()

	// 定义 WorkflowRequest
	var workflowReq = &dify.WorkflowRequest{
		Inputs: map[string]interface{}{
			"input1": "value1",
			"input2": "value2",
		},
		ResponseMode: "blocking", // 设置为阻塞模式以等待返回
		User:         "jiuquan AI",
	}

	// 发送请求并获取响应
	var workflowResp *dify.WorkflowResponse
	if workflowResp, err = client.Api().RunWorkflow(ctx, workflowReq); err != nil {
		t.Fatal(err.Error())
		return
	}

	// 打印响应内容
	j, _ := json.Marshal(workflowResp)
	t.Log(string(j))

	// 验证响应是否符合预期
	if workflowResp.WorkflowRunID == "" || workflowResp.Data.Status == "" {
		t.Fatalf("Invalid workflow response: %v", workflowResp)
	}

	t.Logf("Workflow Run ID: %s", workflowResp.WorkflowRunID)
	t.Logf("Workflow Status: %s", workflowResp.Data.Status)
	t.Logf("Workflow Outputs: %v", workflowResp.Data.Outputs)
}
