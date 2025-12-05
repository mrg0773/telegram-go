package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const (
	defaultTimeout = 30 * time.Second
)

// Client is a Telegram Bot API client wrapper over tgbotapi
type Client struct {
	bot        *tgbotapi.BotAPI
	token      string
	httpClient *http.Client
	logger     *zap.Logger
	debug      bool
}

// Option is a functional option for Client
type Option func(*Client)

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

// WithDebug enables debug mode
func WithDebug(debug bool) Option {
	return func(c *Client) {
		c.debug = debug
	}
}

// NewClient creates a new Telegram client using tgbotapi
func NewClient(token string, logger *zap.Logger, opts ...Option) *Client {
	c := &Client{
		token: token,
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

// initBot lazily initializes the tgbotapi.BotAPI
func (c *Client) initBot() error {
	if c.bot != nil {
		return nil
	}

	bot, err := tgbotapi.NewBotAPIWithClient(c.token, tgbotapi.APIEndpoint, c.httpClient)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	bot.Debug = c.debug
	c.bot = bot
	return nil
}

// GetBot returns the underlying tgbotapi.BotAPI instance
func (c *Client) GetBot() (*tgbotapi.BotAPI, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}
	return c.bot, nil
}

// SendMessage sends a text message to Telegram
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string, opts map[string]interface{}) (*Message, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	msg := tgbotapi.NewMessage(chatID, text)

	// Apply options
	if parseMode, ok := opts["parse_mode"].(string); ok {
		msg.ParseMode = parseMode
	}
	if disablePreview, ok := opts["disable_web_page_preview"].(bool); ok {
		msg.DisableWebPagePreview = disablePreview
	}
	if disableNotification, ok := opts["disable_notification"].(bool); ok {
		msg.DisableNotification = disableNotification
	}
	if replyTo, ok := opts["reply_to_message_id"].(int); ok {
		msg.ReplyToMessageID = replyTo
	}
	if replyMarkup, ok := opts["reply_markup"]; ok {
		msg.ReplyMarkup = replyMarkup
	}

	if c.logger != nil {
		c.logger.Debug("sending message",
			zap.Int64("chat_id", chatID),
			zap.String("text", text),
		)
	}

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return convertMessage(&sent), nil
}

// SendPhoto sends a photo
func (c *Client) SendPhoto(ctx context.Context, chatID int64, photo string, caption string, opts map[string]interface{}) (*Message, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	msg := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(photo))
	msg.Caption = caption

	applyMediaOptions(&msg.BaseChat, &msg.Caption, opts)
	if parseMode, ok := opts["parse_mode"].(string); ok {
		msg.ParseMode = parseMode
	}

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return convertMessage(&sent), nil
}

// SendDocument sends a document
func (c *Client) SendDocument(ctx context.Context, chatID int64, document string, caption string, opts map[string]interface{}) (*Message, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	msg := tgbotapi.NewDocument(chatID, tgbotapi.FileURL(document))
	msg.Caption = caption

	applyMediaOptions(&msg.BaseChat, &msg.Caption, opts)
	if parseMode, ok := opts["parse_mode"].(string); ok {
		msg.ParseMode = parseMode
	}

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return convertMessage(&sent), nil
}

// SendVideo sends a video
func (c *Client) SendVideo(ctx context.Context, chatID int64, video string, caption string, opts map[string]interface{}) (*Message, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	msg := tgbotapi.NewVideo(chatID, tgbotapi.FileURL(video))
	msg.Caption = caption

	applyMediaOptions(&msg.BaseChat, &msg.Caption, opts)
	if parseMode, ok := opts["parse_mode"].(string); ok {
		msg.ParseMode = parseMode
	}

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return convertMessage(&sent), nil
}

// SendAudio sends an audio file
func (c *Client) SendAudio(ctx context.Context, chatID int64, audio string, caption string, opts map[string]interface{}) (*Message, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	msg := tgbotapi.NewAudio(chatID, tgbotapi.FileURL(audio))
	msg.Caption = caption

	applyMediaOptions(&msg.BaseChat, &msg.Caption, opts)
	if parseMode, ok := opts["parse_mode"].(string); ok {
		msg.ParseMode = parseMode
	}

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return convertMessage(&sent), nil
}

// SendVoice sends a voice message
func (c *Client) SendVoice(ctx context.Context, chatID int64, voice string, caption string, opts map[string]interface{}) (*Message, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	msg := tgbotapi.NewVoice(chatID, tgbotapi.FileURL(voice))
	msg.Caption = caption

	applyMediaOptions(&msg.BaseChat, &msg.Caption, opts)
	if parseMode, ok := opts["parse_mode"].(string); ok {
		msg.ParseMode = parseMode
	}

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return convertMessage(&sent), nil
}

// SendVideoNote sends a video note (round video)
func (c *Client) SendVideoNote(ctx context.Context, chatID int64, videoNote string, opts map[string]interface{}) (*Message, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	msg := tgbotapi.NewVideoNote(chatID, 240, tgbotapi.FileURL(videoNote))

	applyBaseOptions(&msg.BaseChat, opts)

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return convertMessage(&sent), nil
}

// SendSticker sends a sticker
func (c *Client) SendSticker(ctx context.Context, chatID int64, sticker string, opts map[string]interface{}) (*Message, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	// Check if sticker is file_id or URL
	var file tgbotapi.RequestFileData
	if len(sticker) > 100 || sticker[0] == 'h' {
		file = tgbotapi.FileURL(sticker)
	} else {
		file = tgbotapi.FileID(sticker)
	}

	msg := tgbotapi.NewSticker(chatID, file)

	applyBaseOptions(&msg.BaseChat, opts)

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return convertMessage(&sent), nil
}

// SendDice sends a dice animation
func (c *Client) SendDice(ctx context.Context, chatID int64, emoji string, opts map[string]interface{}) (*Message, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	msg := tgbotapi.NewDice(chatID)
	msg.Emoji = emoji

	applyBaseOptions(&msg.BaseChat, opts)

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return convertMessage(&sent), nil
}

// SendContact sends a contact
func (c *Client) SendContact(ctx context.Context, chatID int64, contact map[string]interface{}, opts map[string]interface{}) (*Message, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	phoneNumber, _ := contact["phone_number"].(string)
	firstName, _ := contact["first_name"].(string)

	msg := tgbotapi.NewContact(chatID, phoneNumber, firstName)
	if lastName, ok := contact["last_name"].(string); ok {
		msg.LastName = lastName
	}
	if vcard, ok := contact["vcard"].(string); ok {
		msg.VCard = vcard
	}

	applyBaseOptions(&msg.BaseChat, opts)

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return convertMessage(&sent), nil
}

// SendPoll sends a poll
func (c *Client) SendPoll(ctx context.Context, chatID int64, poll map[string]interface{}, opts map[string]interface{}) (*Message, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	question, _ := poll["question"].(string)
	options, _ := poll["options"].([]string)
	if options == nil {
		if optionsRaw, ok := poll["options"].([]interface{}); ok {
			for _, opt := range optionsRaw {
				if s, ok := opt.(string); ok {
					options = append(options, s)
				}
			}
		}
	}

	msg := tgbotapi.NewPoll(chatID, question, options...)

	if isAnonymous, ok := poll["is_anonymous"].(bool); ok {
		msg.IsAnonymous = isAnonymous
	}
	if pollType, ok := poll["type"].(string); ok {
		msg.Type = pollType
	}
	if allowsMultiple, ok := poll["allows_multiple_answers"].(bool); ok {
		msg.AllowsMultipleAnswers = allowsMultiple
	}

	applyBaseOptions(&msg.BaseChat, opts)

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return convertMessage(&sent), nil
}

// SendVenue sends a venue
func (c *Client) SendVenue(ctx context.Context, chatID int64, venue map[string]interface{}, opts map[string]interface{}) (*Message, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	latitude, _ := venue["latitude"].(float64)
	longitude, _ := venue["longitude"].(float64)
	title, _ := venue["title"].(string)
	address, _ := venue["address"].(string)

	msg := tgbotapi.NewVenue(chatID, title, address, latitude, longitude)

	if foursquareID, ok := venue["foursquare_id"].(string); ok {
		msg.FoursquareID = foursquareID
	}
	if foursquareType, ok := venue["foursquare_type"].(string); ok {
		msg.FoursquareType = foursquareType
	}

	applyBaseOptions(&msg.BaseChat, opts)

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return convertMessage(&sent), nil
}

// SendLocation sends a location
func (c *Client) SendLocation(ctx context.Context, chatID int64, latitude, longitude float64, opts map[string]interface{}) (*Message, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	msg := tgbotapi.NewLocation(chatID, latitude, longitude)

	applyBaseOptions(&msg.BaseChat, opts)

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return convertMessage(&sent), nil
}

// SendGame sends a game
func (c *Client) SendGame(ctx context.Context, chatID int64, gameShortName string, opts map[string]interface{}) (*Message, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	msg := tgbotapi.GameConfig{
		BaseChat:      tgbotapi.BaseChat{ChatID: chatID},
		GameShortName: gameShortName,
	}

	applyBaseOptions(&msg.BaseChat, opts)

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return convertMessage(&sent), nil
}

// SendChatAction sends a chat action (typing, upload_photo, etc.)
func (c *Client) SendChatAction(ctx context.Context, chatID int64, action string) error {
	if err := c.initBot(); err != nil {
		return err
	}

	msg := tgbotapi.NewChatAction(chatID, action)
	_, err := c.bot.Request(msg)
	return c.wrapError(err)
}

// EditMessageText edits text of a message
func (c *Client) EditMessageText(ctx context.Context, chatID int64, messageID int64, text string, opts map[string]interface{}) (*Message, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	msg := tgbotapi.NewEditMessageText(chatID, int(messageID), text)

	if parseMode, ok := opts["parse_mode"].(string); ok {
		msg.ParseMode = parseMode
	}
	if disablePreview, ok := opts["disable_web_page_preview"].(bool); ok {
		msg.DisableWebPagePreview = disablePreview
	}
	if replyMarkup, ok := opts["reply_markup"].(tgbotapi.InlineKeyboardMarkup); ok {
		msg.ReplyMarkup = &replyMarkup
	}

	sent, err := c.bot.Send(msg)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return convertMessage(&sent), nil
}

// DeleteMessage deletes a message
func (c *Client) DeleteMessage(ctx context.Context, chatID int64, messageID int64) error {
	if err := c.initBot(); err != nil {
		return err
	}

	msg := tgbotapi.NewDeleteMessage(chatID, int(messageID))
	_, err := c.bot.Request(msg)
	return c.wrapError(err)
}

// AnswerCallbackQuery answers a callback query
func (c *Client) AnswerCallbackQuery(ctx context.Context, callbackQueryID string, opts map[string]interface{}) error {
	if err := c.initBot(); err != nil {
		return err
	}

	callback := tgbotapi.NewCallback(callbackQueryID, "")

	if text, ok := opts["text"].(string); ok {
		callback.Text = text
	}
	if showAlert, ok := opts["show_alert"].(bool); ok {
		callback.ShowAlert = showAlert
	}
	if url, ok := opts["url"].(string); ok {
		callback.URL = url
	}
	if cacheTime, ok := opts["cache_time"].(int); ok {
		callback.CacheTime = cacheTime
	}

	_, err := c.bot.Request(callback)
	return c.wrapError(err)
}

// GetFile gets file info by file_id
func (c *Client) GetFile(ctx context.Context, fileID string) (*FileResponse, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	file, err := c.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, c.wrapError(err)
	}

	return &FileResponse{
		FileID:       file.FileID,
		FileUniqueID: file.FileUniqueID,
		FileSize:     int64(file.FileSize),
		FilePath:     file.FilePath,
	}, nil
}

// GetFileURL returns URL to download file
func (c *Client) GetFileURL(filePath string) string {
	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", c.token, filePath)
}

// SetWebhook sets webhook URL
func (c *Client) SetWebhook(ctx context.Context, url string, opts map[string]interface{}) error {
	if err := c.initBot(); err != nil {
		return err
	}

	webhook, err := tgbotapi.NewWebhook(url)
	if err != nil {
		return err
	}

	if maxConnections, ok := opts["max_connections"].(int); ok {
		webhook.MaxConnections = maxConnections
	}

	_, err = c.bot.Request(webhook)
	return c.wrapError(err)
}

// DeleteWebhook deletes webhook
func (c *Client) DeleteWebhook(ctx context.Context, dropPending bool) error {
	if err := c.initBot(); err != nil {
		return err
	}

	_, err := c.bot.Request(tgbotapi.DeleteWebhookConfig{
		DropPendingUpdates: dropPending,
	})
	return c.wrapError(err)
}

// GetMe returns bot info
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	user, err := c.bot.GetMe()
	if err != nil {
		return nil, c.wrapError(err)
	}

	return &User{
		ID:           user.ID,
		IsBot:        user.IsBot,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Username:     user.UserName,
		LanguageCode: user.LanguageCode,
	}, nil
}

// Call makes a raw API call with any method and parameters
// This method exists for backward compatibility
func (c *Client) Call(ctx context.Context, method string, params map[string]interface{}) (*Response, error) {
	if err := c.initBot(); err != nil {
		return nil, err
	}

	// Convert params to JSON for tgbotapi Params
	tgParams := make(tgbotapi.Params)
	for k, v := range params {
		switch val := v.(type) {
		case string:
			tgParams[k] = val
		case int:
			tgParams.AddNonZero(k, val)
		case int64:
			tgParams.AddNonZero64(k, val)
		case float64:
			tgParams.AddNonZeroFloat(k, val)
		case bool:
			tgParams.AddBool(k, val)
		default:
			// For complex types, marshal to JSON
			jsonBytes, err := json.Marshal(val)
			if err == nil {
				tgParams[k] = string(jsonBytes)
			}
		}
	}

	resp, err := c.bot.MakeRequest(method, tgParams)
	if err != nil {
		return nil, c.wrapError(err)
	}

	return &Response{
		OK:          resp.Ok,
		Result:      resp.Result,
		Description: resp.Description,
		ErrorCode:   resp.ErrorCode,
	}, nil
}

// wrapError converts tgbotapi errors to APIError
func (c *Client) wrapError(err error) error {
	if err == nil {
		return nil
	}

	// Try to extract error code from tgbotapi error
	if tgErr, ok := err.(*tgbotapi.Error); ok {
		return &APIError{
			Code:        tgErr.Code,
			Description: tgErr.Message,
		}
	}

	return err
}

// Helper functions

func applyBaseOptions(base *tgbotapi.BaseChat, opts map[string]interface{}) {
	if disableNotification, ok := opts["disable_notification"].(bool); ok {
		base.DisableNotification = disableNotification
	}
	if replyTo, ok := opts["reply_to_message_id"].(int); ok {
		base.ReplyToMessageID = replyTo
	}
	if replyMarkup, ok := opts["reply_markup"]; ok {
		base.ReplyMarkup = replyMarkup
	}
}

func applyMediaOptions(base *tgbotapi.BaseChat, caption *string, opts map[string]interface{}) {
	applyBaseOptions(base, opts)
}

// convertMessage converts tgbotapi.Message to our Message type
func convertMessage(msg *tgbotapi.Message) *Message {
	if msg == nil {
		return nil
	}

	result := &Message{
		MessageID: int64(msg.MessageID),
		Date:      int64(msg.Date),
		Text:      msg.Text,
		Caption:   msg.Caption,
		Chat: Chat{
			ID:        msg.Chat.ID,
			Type:      msg.Chat.Type,
			Title:     msg.Chat.Title,
			Username:  msg.Chat.UserName,
			FirstName: msg.Chat.FirstName,
			LastName:  msg.Chat.LastName,
		},
	}

	if msg.From != nil {
		result.From = &User{
			ID:           msg.From.ID,
			IsBot:        msg.From.IsBot,
			FirstName:    msg.From.FirstName,
			LastName:     msg.From.LastName,
			Username:     msg.From.UserName,
			LanguageCode: msg.From.LanguageCode,
		}
	}

	if msg.ReplyToMessage != nil {
		result.ReplyToMessage = convertMessage(msg.ReplyToMessage)
	}

	// Convert photo
	if msg.Photo != nil {
		for _, p := range msg.Photo {
			result.Photo = append(result.Photo, PhotoSize{
				FileID:       p.FileID,
				FileUniqueID: p.FileUniqueID,
				Width:        p.Width,
				Height:       p.Height,
				FileSize:     int64(p.FileSize),
			})
		}
	}

	// Convert document
	if msg.Document != nil {
		result.Document = &Document{
			FileID:       msg.Document.FileID,
			FileUniqueID: msg.Document.FileUniqueID,
			FileName:     msg.Document.FileName,
			MimeType:     msg.Document.MimeType,
			FileSize:     int64(msg.Document.FileSize),
		}
	}

	// Convert video
	if msg.Video != nil {
		result.Video = &Video{
			FileID:       msg.Video.FileID,
			FileUniqueID: msg.Video.FileUniqueID,
			Width:        msg.Video.Width,
			Height:       msg.Video.Height,
			Duration:     msg.Video.Duration,
			FileName:     msg.Video.FileName,
			MimeType:     msg.Video.MimeType,
			FileSize:     int64(msg.Video.FileSize),
		}
	}

	// Convert audio
	if msg.Audio != nil {
		result.Audio = &Audio{
			FileID:       msg.Audio.FileID,
			FileUniqueID: msg.Audio.FileUniqueID,
			Duration:     msg.Audio.Duration,
			Performer:    msg.Audio.Performer,
			Title:        msg.Audio.Title,
			FileName:     msg.Audio.FileName,
			MimeType:     msg.Audio.MimeType,
			FileSize:     int64(msg.Audio.FileSize),
		}
	}

	// Convert voice
	if msg.Voice != nil {
		result.Voice = &Voice{
			FileID:       msg.Voice.FileID,
			FileUniqueID: msg.Voice.FileUniqueID,
			Duration:     msg.Voice.Duration,
			MimeType:     msg.Voice.MimeType,
			FileSize:     int64(msg.Voice.FileSize),
		}
	}

	// Convert sticker
	if msg.Sticker != nil {
		result.Sticker = &Sticker{
			FileID:       msg.Sticker.FileID,
			FileUniqueID: msg.Sticker.FileUniqueID,
			Width:        msg.Sticker.Width,
			Height:       msg.Sticker.Height,
			IsAnimated:   msg.Sticker.IsAnimated,
			Emoji:        msg.Sticker.Emoji,
			SetName:      msg.Sticker.SetName,
			FileSize:     int64(msg.Sticker.FileSize),
		}
	}

	// Convert contact
	if msg.Contact != nil {
		result.Contact = &Contact{
			PhoneNumber: msg.Contact.PhoneNumber,
			FirstName:   msg.Contact.FirstName,
			LastName:    msg.Contact.LastName,
			UserID:      msg.Contact.UserID,
			VCard:       msg.Contact.VCard,
		}
	}

	// Convert location
	if msg.Location != nil {
		result.Location = &Location{
			Longitude: msg.Location.Longitude,
			Latitude:  msg.Location.Latitude,
		}
	}

	// Convert venue
	if msg.Venue != nil {
		result.Venue = &Venue{
			Location: Location{
				Longitude: msg.Venue.Location.Longitude,
				Latitude:  msg.Venue.Location.Latitude,
			},
			Title:          msg.Venue.Title,
			Address:        msg.Venue.Address,
			FoursquareID:   msg.Venue.FoursquareID,
			FoursquareType: msg.Venue.FoursquareType,
		}
	}

	// Convert poll
	if msg.Poll != nil {
		result.Poll = &Poll{
			ID:                    msg.Poll.ID,
			Question:              msg.Poll.Question,
			TotalVoterCount:       msg.Poll.TotalVoterCount,
			IsClosed:              msg.Poll.IsClosed,
			IsAnonymous:           msg.Poll.IsAnonymous,
			Type:                  msg.Poll.Type,
			AllowsMultipleAnswers: msg.Poll.AllowsMultipleAnswers,
		}
		for _, opt := range msg.Poll.Options {
			result.Poll.Options = append(result.Poll.Options, PollOption{
				Text:       opt.Text,
				VoterCount: opt.VoterCount,
			})
		}
	}

	// Convert dice
	if msg.Dice != nil {
		result.Dice = &Dice{
			Emoji: msg.Dice.Emoji,
			Value: msg.Dice.Value,
		}
	}

	return result
}
