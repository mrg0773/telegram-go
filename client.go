package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const (
	defaultBaseURL = "https://api.telegram.org/bot"
	defaultTimeout = 30 * time.Second
)

// Client is a Telegram Bot API client
type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// Option is a functional option for Client
type Option func(*Client)

// WithBaseURL sets custom base URL (for testing)
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithTimeout sets custom HTTP timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithHTTPClient sets custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// NewClient creates a new Telegram client
func NewClient(token string, logger *zap.Logger, opts ...Option) *Client {
	c := &Client{
		token:   token,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		logger: logger,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// SendMessage sends a text message to Telegram
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string, opts map[string]interface{}) (*Message, error) {
	params := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	for k, v := range opts {
		params[k] = v
	}

	return c.sendMethod(ctx, "sendMessage", params)
}

// SendPhoto sends a photo
func (c *Client) SendPhoto(ctx context.Context, chatID int64, photo string, caption string, opts map[string]interface{}) (*Message, error) {
	params := map[string]interface{}{
		"chat_id": chatID,
		"photo":   photo,
	}
	if caption != "" {
		params["caption"] = caption
	}
	for k, v := range opts {
		params[k] = v
	}
	return c.sendMethod(ctx, "sendPhoto", params)
}

// SendDocument sends a document
func (c *Client) SendDocument(ctx context.Context, chatID int64, document string, caption string, opts map[string]interface{}) (*Message, error) {
	params := map[string]interface{}{
		"chat_id":  chatID,
		"document": document,
	}
	if caption != "" {
		params["caption"] = caption
	}
	for k, v := range opts {
		params[k] = v
	}
	return c.sendMethod(ctx, "sendDocument", params)
}

// SendVideo sends a video
func (c *Client) SendVideo(ctx context.Context, chatID int64, video string, caption string, opts map[string]interface{}) (*Message, error) {
	params := map[string]interface{}{
		"chat_id": chatID,
		"video":   video,
	}
	if caption != "" {
		params["caption"] = caption
	}
	for k, v := range opts {
		params[k] = v
	}
	return c.sendMethod(ctx, "sendVideo", params)
}

// SendAudio sends an audio file
func (c *Client) SendAudio(ctx context.Context, chatID int64, audio string, caption string, opts map[string]interface{}) (*Message, error) {
	params := map[string]interface{}{
		"chat_id": chatID,
		"audio":   audio,
	}
	if caption != "" {
		params["caption"] = caption
	}
	for k, v := range opts {
		params[k] = v
	}
	return c.sendMethod(ctx, "sendAudio", params)
}

// SendVoice sends a voice message
func (c *Client) SendVoice(ctx context.Context, chatID int64, voice string, caption string, opts map[string]interface{}) (*Message, error) {
	params := map[string]interface{}{
		"chat_id": chatID,
		"voice":   voice,
	}
	if caption != "" {
		params["caption"] = caption
	}
	for k, v := range opts {
		params[k] = v
	}
	return c.sendMethod(ctx, "sendVoice", params)
}

// SendVideoNote sends a video note (round video)
func (c *Client) SendVideoNote(ctx context.Context, chatID int64, videoNote string, opts map[string]interface{}) (*Message, error) {
	params := map[string]interface{}{
		"chat_id":    chatID,
		"video_note": videoNote,
	}
	for k, v := range opts {
		params[k] = v
	}
	return c.sendMethod(ctx, "sendVideoNote", params)
}

// SendSticker sends a sticker
func (c *Client) SendSticker(ctx context.Context, chatID int64, sticker string, opts map[string]interface{}) (*Message, error) {
	params := map[string]interface{}{
		"chat_id": chatID,
		"sticker": sticker,
	}
	for k, v := range opts {
		params[k] = v
	}
	return c.sendMethod(ctx, "sendSticker", params)
}

// SendDice sends a dice animation
func (c *Client) SendDice(ctx context.Context, chatID int64, emoji string, opts map[string]interface{}) (*Message, error) {
	params := map[string]interface{}{
		"chat_id": chatID,
		"emoji":   emoji,
	}
	for k, v := range opts {
		params[k] = v
	}
	return c.sendMethod(ctx, "sendDice", params)
}

// SendContact sends a contact
func (c *Client) SendContact(ctx context.Context, chatID int64, contact map[string]interface{}, opts map[string]interface{}) (*Message, error) {
	params := map[string]interface{}{
		"chat_id": chatID,
	}
	for k, v := range contact {
		params[k] = v
	}
	for k, v := range opts {
		params[k] = v
	}
	return c.sendMethod(ctx, "sendContact", params)
}

// SendPoll sends a poll
func (c *Client) SendPoll(ctx context.Context, chatID int64, poll map[string]interface{}, opts map[string]interface{}) (*Message, error) {
	params := map[string]interface{}{
		"chat_id": chatID,
	}
	for k, v := range poll {
		params[k] = v
	}
	for k, v := range opts {
		params[k] = v
	}
	return c.sendMethod(ctx, "sendPoll", params)
}

// SendVenue sends a venue
func (c *Client) SendVenue(ctx context.Context, chatID int64, venue map[string]interface{}, opts map[string]interface{}) (*Message, error) {
	params := map[string]interface{}{
		"chat_id": chatID,
	}
	for k, v := range venue {
		params[k] = v
	}
	for k, v := range opts {
		params[k] = v
	}
	return c.sendMethod(ctx, "sendVenue", params)
}

// SendLocation sends a location
func (c *Client) SendLocation(ctx context.Context, chatID int64, latitude, longitude float64, opts map[string]interface{}) (*Message, error) {
	params := map[string]interface{}{
		"chat_id":   chatID,
		"latitude":  latitude,
		"longitude": longitude,
	}
	for k, v := range opts {
		params[k] = v
	}
	return c.sendMethod(ctx, "sendLocation", params)
}

// SendGame sends a game
func (c *Client) SendGame(ctx context.Context, chatID int64, gameShortName string, opts map[string]interface{}) (*Message, error) {
	params := map[string]interface{}{
		"chat_id":         chatID,
		"game_short_name": gameShortName,
	}
	for k, v := range opts {
		params[k] = v
	}
	return c.sendMethod(ctx, "sendGame", params)
}

// SendChatAction sends a chat action (typing, upload_photo, etc.)
func (c *Client) SendChatAction(ctx context.Context, chatID int64, action string) error {
	params := map[string]interface{}{
		"chat_id": chatID,
		"action":  action,
	}
	_, err := c.request(ctx, "sendChatAction", params)
	return err
}

// EditMessageText edits text of a message
func (c *Client) EditMessageText(ctx context.Context, chatID int64, messageID int64, text string, opts map[string]interface{}) (*Message, error) {
	params := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
	}
	for k, v := range opts {
		params[k] = v
	}
	return c.sendMethod(ctx, "editMessageText", params)
}

// DeleteMessage deletes a message
func (c *Client) DeleteMessage(ctx context.Context, chatID int64, messageID int64) error {
	params := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
	}
	_, err := c.request(ctx, "deleteMessage", params)
	return err
}

// AnswerCallbackQuery answers a callback query
func (c *Client) AnswerCallbackQuery(ctx context.Context, callbackQueryID string, opts map[string]interface{}) error {
	params := map[string]interface{}{
		"callback_query_id": callbackQueryID,
	}
	for k, v := range opts {
		params[k] = v
	}
	_, err := c.request(ctx, "answerCallbackQuery", params)
	return err
}

// GetFile gets file info by file_id
func (c *Client) GetFile(ctx context.Context, fileID string) (*FileResponse, error) {
	params := map[string]interface{}{
		"file_id": fileID,
	}

	resp, err := c.request(ctx, "getFile", params)
	if err != nil {
		return nil, err
	}

	var file FileResponse
	if err := json.Unmarshal(resp.Result, &file); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file response: %w", err)
	}

	return &file, nil
}

// GetFileURL returns URL to download file
func (c *Client) GetFileURL(filePath string) string {
	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", c.token, filePath)
}

// SetWebhook sets webhook URL
func (c *Client) SetWebhook(ctx context.Context, url string, opts map[string]interface{}) error {
	params := map[string]interface{}{
		"url": url,
	}
	for k, v := range opts {
		params[k] = v
	}
	_, err := c.request(ctx, "setWebhook", params)
	return err
}

// DeleteWebhook deletes webhook
func (c *Client) DeleteWebhook(ctx context.Context, dropPending bool) error {
	params := map[string]interface{}{
		"drop_pending_updates": dropPending,
	}
	_, err := c.request(ctx, "deleteWebhook", params)
	return err
}

// GetMe returns bot info
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	resp, err := c.request(ctx, "getMe", nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(resp.Result, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// Call makes a raw API call with any method and parameters
func (c *Client) Call(ctx context.Context, method string, params map[string]interface{}) (*Response, error) {
	return c.request(ctx, method, params)
}

// sendMethod is a helper that sends a message and returns Message
func (c *Client) sendMethod(ctx context.Context, method string, params map[string]interface{}) (*Message, error) {
	resp, err := c.requestWithRetry(ctx, method, params, 3)
	if err != nil {
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal(resp.Result, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}

// request makes an API request
func (c *Client) request(ctx context.Context, method string, params map[string]interface{}) (*Response, error) {
	url := fmt.Sprintf("%s%s/%s", c.baseURL, c.token, method)

	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if c.logger != nil {
		c.logger.Debug("telegram request",
			zap.String("method", method),
			zap.String("url", url),
			zap.ByteString("body", body),
		)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if c.logger != nil {
		c.logger.Debug("telegram response",
			zap.String("method", method),
			zap.Int("status", resp.StatusCode),
			zap.ByteString("body", respBody),
		)
	}

	var tgResp Response
	if err := json.Unmarshal(respBody, &tgResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !tgResp.OK {
		return &tgResp, &APIError{
			Code:        tgResp.ErrorCode,
			Description: tgResp.Description,
		}
	}

	return &tgResp, nil
}

// requestWithRetry makes an API request with retry on rate limit
func (c *Client) requestWithRetry(ctx context.Context, method string, params map[string]interface{}, maxRetries int) (*Response, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		resp, err := c.request(ctx, method, params)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Check if it's a rate limit error (429)
		if apiErr, ok := err.(*APIError); ok && apiErr.Code == 429 {
			// Wait before retry with exponential backoff
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		// For other errors, return immediately
		return resp, err
	}
	return nil, lastErr
}
