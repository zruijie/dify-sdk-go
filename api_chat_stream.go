package dify

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type ChatMessageStreamResponse struct {
	Event          string `json:"event"`
	TaskID         string `json:"task_id"`
	ID             string `json:"id"`
	Answer         string `json:"answer"`
	CreatedAt      int64  `json:"created_at"`
	ConversationID string `json:"conversation_id"`
}

type ChatMessageStreamChannelResponse struct {
	ChatMessageStreamResponse
	Err error `json:"-"`
}

func (api *API) ChatMessagesStreamRaw(ctx context.Context, req *ChatMessageRequest) (*http.Response, error) {
	req.ResponseMode = "streaming"

	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, "/v1/chat-messages", req)
	if err != nil {
		return nil, err
	}
	return api.c.sendRequest(httpReq)
}

func (api *API) ChatMessagesStream(ctx context.Context, req *ChatMessageRequest) (chan ChatMessageStreamChannelResponse, error) {
	httpResp, err := api.ChatMessagesStreamRaw(ctx, req)
	if err != nil {
		return nil, err
	}

	streamChannel := make(chan ChatMessageStreamChannelResponse)
	go api.chatMessagesStreamHandle(ctx, httpResp, streamChannel)
	return streamChannel, nil
}

func (api *API) chatMessagesStreamHandle(ctx context.Context, resp *http.Response, streamChannel chan ChatMessageStreamChannelResponse) {
	var (
		body   = resp.Body
		reader = bufio.NewReader(body)

		err  error
		line []byte
	)

	defer resp.Body.Close()
	defer close(streamChannel)

	if line, _, err = reader.ReadLine(); err == nil {
		var errResp ErrorResponse
		var _err error
		if _err = json.Unmarshal(line, &errResp); _err == nil {
			streamChannel <- ChatMessageStreamChannelResponse{
				Err: errors.New(string(line)),
			}
			return
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if line, err = reader.ReadBytes('\n'); err != nil {
				streamChannel <- ChatMessageStreamChannelResponse{
					Err: errors.New("Error reading line: " + err.Error()),
				}
				return
			}

			if !bytes.HasPrefix(line, []byte("data:")) {
				continue
			}

			line = bytes.TrimPrefix(line, []byte("data:"))
			line = bytes.TrimSpace(line)

			var resp ChatMessageStreamChannelResponse
			if err = json.Unmarshal(line, &resp); err != nil {
				streamChannel <- ChatMessageStreamChannelResponse{
					Err: errors.New("Error unmarshalling event: " + err.Error()),
				}
				return
			} else if resp.Answer == "" {
				return
			}
			streamChannel <- resp
		}
	}
}
