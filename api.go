package dify

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var (
	chatMessages        = "/v1/chat-messages"
	messagesFeedbacks   = "/v1/messages/{message_id}/feedbacks"
	messages            = "/v1/messages"
	conversations       = "/v1/conversations"
	conversationsRename = "/v1/conversations/{conversation_id}/name"
	parameters          = "/v1/parameters"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Params  string `json:"params"`
}

type ChatMessageRequest struct {
	Inputs         map[string]interface{} `json:"inputs"`
	Query          string                 `json:"query"`
	ResponseMode   string                 `json:"response_mode"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	User           string                 `json:"user"`
}

type ChatMessageResponse struct {
	ID             string `json:"id"`
	Answer         string `json:"answer"`
	ConversationID string `json:"conversation_id"`
	CreatedAt      int    `json:"created_at"`
}

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

type MessagesFeedbacksRequest struct {
	MessageID string `json:"message_id,omitempty"`
	Rating    string `json:"rating,omitempty"`
	User      string `json:"user"`
}

type MessagesFeedbacksResponse struct {
	HasMore bool                            `json:"has_more"`
	Data    []MessagesFeedbacksDataResponse `json:"data"`
}

type MessagesFeedbacksDataResponse struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	PhoneNumber    string `json:"phone_number"`
	AvatarURL      string `json:"avatar_url"`
	DisplayName    string `json:"display_name"`
	ConversationID string `json:"conversation_id"`
	LastActiveAt   int64  `json:"last_active_at"`
	CreatedAt      int64  `json:"created_at"`
}

type MessagesRequest struct {
	ConversationID string `json:"conversation_id"`
	FirstID        string `json:"first_id,omitempty"`
	Limit          int    `json:"limit"`
	User           string `json:"user"`
}

type MessagesResponse struct {
	Limit   int                    `json:"limit"`
	HasMore bool                   `json:"has_more"`
	Data    []MessagesDataResponse `json:"data"`
}

type MessagesDataResponse struct {
	ID             string                 `json:"id"`
	ConversationID string                 `json:"conversation_id"`
	Inputs         map[string]interface{} `json:"inputs"`
	Query          string                 `json:"query"`
	Answer         string                 `json:"answer"`
	Feedback       interface{}            `json:"feedback"`
	CreatedAt      int64                  `json:"created_at"`
}

type ConversationsRequest struct {
	LastID string `json:"last_id,omitempty"`
	Limit  int    `json:"limit"`
	User   string `json:"user"`
}

type ConversationsResponse struct {
	Limit   int                         `json:"limit"`
	HasMore bool                        `json:"has_more"`
	Data    []ConversationsDataResponse `json:"data"`
}

type ConversationsDataResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Inputs    map[string]string `json:"inputs"`
	Status    string            `json:"status"`
	CreatedAt int64             `json:"created_at"`
}

type ConversationsRenamingRequest struct {
	ConversationID string `json:"conversation_id,omitempty"`
	Name           string `json:"name"`
	User           string `json:"user"`
}

type ConversationsRenamingResponse struct {
	Result string `json:"result"`
}

type ParametersRequest struct {
	User string `json:"user"`
}

type ParametersResponse struct {
	OpeningStatement              string        `json:"opening_statement"`
	SuggestedQuestions            []interface{} `json:"suggested_questions"`
	SuggestedQuestionsAfterAnswer struct {
		Enabled bool `json:"enabled"`
	} `json:"suggested_questions_after_answer"`
	MoreLikeThis struct {
		Enabled bool `json:"enabled"`
	} `json:"more_like_this"`
	UserInputForm []map[string]interface{} `json:"user_input_form"`
}

// type ParametersUserInputFormResponse struct {
// 	TextInput []ParametersTextInputResponse `json:"text-input"`
// }

// type ParametersTextInputResponse struct {
// 	Label     string `json:"label"`
// 	Variable  string `json:"variable"`
// 	Required  bool   `json:"required"`
// 	MaxLength int    `json:"max_length"`
// 	Default   string `json:"default"`
// }

const (
	FeedbackLike    = "like"
	FeedbackDislike = "dislike"
)

type API struct {
	c         *Client
	apiSecret string
}

func (api *API) WithAPISecret(apiSecret string) *API {
	api.apiSecret = apiSecret
	return api
}

func (api *API) getApiSecretKey() string {
	if api.apiSecret != "" {
		return api.apiSecret
	}
	return api.c.getApiSecretKey()
}

func (api *API) createBaseRequest(ctx context.Context, method, apiUrl string, body interface{}) (*http.Request, error) {
	var b io.Reader
	if body != nil {
		reqBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		b = bytes.NewBuffer(reqBytes)
	} else {
		b = http.NoBody
	}
	req, err := http.NewRequestWithContext(ctx, method, api.c.getHost()+apiUrl, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+api.getApiSecretKey())
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	return req, nil
}

/* Create chat message
 * Create a new conversation message or continue an existing dialogue.
 */
func (api *API) ChatMessages(ctx context.Context, req *ChatMessageRequest) (resp *ChatMessageResponse, err error) {
	req.ResponseMode = "blocking"

	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, chatMessages, req)
	if err != nil {
		return
	}
	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}

func (api *API) ChatMessagesStreamRaw(ctx context.Context, req *ChatMessageRequest) (*http.Response, error) {
	req.ResponseMode = "streaming"

	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, chatMessages, req)
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

/* Message terminal user feedback, like
 * Rate received messages on behalf of end-users with likes or dislikes.
 * This data is visible in the Logs & Annotations page and used for future model fine-tuning.
 */
func (api *API) MessagesFeedbacks(ctx context.Context, req *MessagesFeedbacksRequest) (resp *MessagesFeedbacksResponse, err error) {
	if req.MessageID == "" {
		err = errors.New("MessagesFeedbacksRequest.MessageID Illegal")
		return
	}

	url := strings.ReplaceAll(messagesFeedbacks, "{message_id}", req.MessageID)
	req.MessageID = ""

	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return
	}
	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}

/* Get the chat history message
 * The first page returns the latest limit bar, which is in reverse order.
 */
func (api *API) Messages(ctx context.Context, req *MessagesRequest) (resp *MessagesResponse, err error) {
	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, messages, nil)
	if err != nil {
		return
	}
	query := httpReq.URL.Query()
	query.Set("conversation_id", req.ConversationID)
	query.Set("user", req.User)
	if req.FirstID != "" {
		query.Set("first_id", req.FirstID)
	}
	if req.Limit > 0 {
		query.Set("limit", strconv.FormatInt(int64(req.Limit), 10))
	}
	httpReq.URL.RawQuery = query.Encode()

	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}

/* Get conversation list
 * Gets the session list of the current user. By default, the last 20 sessions are returned.
 */
func (api *API) Conversations(ctx context.Context, req *ConversationsRequest) (resp *ConversationsResponse, err error) {
	if req.User == "" {
		err = errors.New("ConversationsRequest.User Illegal")
		return
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, conversations, nil)
	if err != nil {
		return
	}

	query := httpReq.URL.Query()
	query.Set("last_id", req.LastID)
	query.Set("user", req.User)
	query.Set("limit", strconv.FormatInt(int64(req.Limit), 10))
	httpReq.URL.RawQuery = query.Encode()

	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}

/* Conversation renaming
 * Rename conversations; the name is displayed in multi-session client interfaces.
 */
func (api *API) ConversationsRenaming(ctx context.Context, req *ConversationsRenamingRequest) (resp *ConversationsRenamingResponse, err error) {
	url := strings.ReplaceAll(conversationsRename, "{conversation_id}", req.ConversationID)
	req.ConversationID = ""

	httpReq, err := api.createBaseRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return
	}
	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}

/* Obtain application parameter information
 * Retrieve configured Input parameters, including variable names, field names, types, and default values.
 * Typically used for displaying these fields in a form or filling in default values after the client loads.
 */
func (api *API) Parameters(ctx context.Context, req *ParametersRequest) (resp *ParametersResponse, err error) {
	if req.User == "" {
		err = errors.New("ParametersRequest.User Illegal")
		return
	}

	httpReq, err := api.createBaseRequest(ctx, http.MethodGet, parameters, nil)
	if err != nil {
		return
	}
	query := httpReq.URL.Query()
	query.Set("user", req.User)
	httpReq.URL.RawQuery = query.Encode()

	err = api.c.sendJSONRequest(httpReq, &resp)
	return
}
