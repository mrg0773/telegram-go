package telegram

import (
	"context"
	"encoding/json"
	"math"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Action represents a message action to execute
// This is the main entry point for sending messages from handler
type Action struct {
	Activity string     `json:"activity,omitempty"` // "message"
	Project  string     `json:"slag,omitempty"`     // Project slug
	User     ActionUser `json:"user,omitempty"`     // User info (TgID required)
	Content  Content    `json:"content,omitempty"`  // Message content
	Token    string     `json:"-"`                  // Bot token (passed separately)
}

// ActionUser represents user information for action
type ActionUser struct {
	TgID int64  `json:"tg_id,omitempty"` // Telegram user ID (chat_id)
	ID   string `json:"id,omitempty"`    // Internal user ID
}

// Content represents the message content
type Content struct {
	Type        string                 `json:"type,omitempty"`         // text, inline_keyboard, virtual_keyboard, sticker, dice, etc.
	Stream      string                 `json:"stream,omitempty"`       // tg_direct
	Text        string                 `json:"text,omitempty"`         // Message text
	Attachment  *Attachment            `json:"attachment,omitempty"`   // Media attachment
	Buts        []string               `json:"buts,omitempty"`         // Button labels
	Actions     []json.RawMessage      `json:"actions,omitempty"`      // Button callback actions
	ReplyMarkup map[string]interface{} `json:"reply_markup,omitempty"` // Custom reply markup
	ColumnNum   *int                   `json:"column_num,omitempty"`   // Keyboard column count
	Spices      map[string]interface{} `json:"spices,omitempty"`       // Extra params (parse_mode, etc.)
	Parameters  Parameters             `json:"parameters,omitempty"`   // Action parameters
}

// Attachment represents media attachment
type Attachment struct {
	Type          string      `json:"type,omitempty"`            // photo, document, video, audio, voice, video_note
	URL           string      `json:"url,omitempty"`             // File URL or file_id
	Sticker       string      `json:"sticker,omitempty"`         // Sticker file_id
	Dice          string      `json:"dice,omitempty"`            // Dice emoji
	Contact       interface{} `json:"contact,omitempty"`         // Contact data
	Poll          interface{} `json:"poll,omitempty"`            // Poll data
	Venue         interface{} `json:"venue,omitempty"`           // Venue data
	GameShortName string      `json:"game_short_name,omitempty"` // Game short name
}

// Parameters represents action parameters
type Parameters struct {
	Save         *bool   `json:"save,omitempty"`          // Save to outbox
	SendReaction *string `json:"send_reaction,omitempty"` // Chat action before send
}

// ActionResult represents the result of action execution
type ActionResult struct {
	Success   bool      `json:"success"`
	MessageID int64     `json:"message_id,omitempty"`
	Response  *Response `json:"response,omitempty"`
	Error     error     `json:"error,omitempty"`
}

// CallbackData represents callback query data for keyboard buttons
type CallbackData struct {
	Project   string          `json:"project"`
	UserID    string          `json:"user_id"`
	QueryData string          `json:"query_data"`
	Action    json.RawMessage `json:"action,omitempty"`
}

// CallbackSaver interface for saving callback data to database
type CallbackSaver interface {
	SaveCallbackData(ctx context.Context, data *CallbackData) error
	SaveCallbackDataBatch(ctx context.Context, data []*CallbackData) error
}

// ExecuteAction executes a message action using tgbotapi
// Returns ActionResult with message ID on success or error on failure
func (c *Client) ExecuteAction(ctx context.Context, action *Action, callbackSaver CallbackSaver) (*ActionResult, error) {
	if action.Content.Stream != "tg_direct" && action.Content.Stream != "" {
		// Only tg_direct stream is supported
		return &ActionResult{Success: false}, nil
	}

	if err := c.initBot(); err != nil {
		return &ActionResult{Success: false, Error: err}, err
	}

	// Apply text formatting
	text := action.Content.Text
	parseMode := ""
	if pm, ok := action.Content.Spices["parse_mode"].(string); ok {
		parseMode = pm
		if parseMode == "MarkdownV2" {
			text = FormatMarkdownV2(text)
		}
	}

	// Send chat action if configured
	if action.Content.Parameters.SendReaction != nil {
		chatAction := tgbotapi.NewChatAction(action.User.TgID, *action.Content.Parameters.SendReaction)
		_, _ = c.bot.Request(chatAction)
	}

	// Build and send message based on content type
	var sent tgbotapi.Message
	var err error

	switch action.Content.Type {
	case "sticker":
		sent, err = c.sendStickerAction(action)
	case "dice":
		sent, err = c.sendDiceAction(action)
	case "contact":
		sent, err = c.sendContactAction(action)
	case "poll":
		sent, err = c.sendPollAction(action, parseMode)
	case "game":
		sent, err = c.sendGameAction(action)
	case "venue":
		sent, err = c.sendVenueAction(action)
	default:
		// Text-based messages (text, inline_keyboard, virtual_keyboard, or empty)
		sent, err = c.sendTextBasedAction(ctx, action, text, parseMode, callbackSaver)
	}

	if err != nil {
		return &ActionResult{Success: false, Error: err}, err
	}

	return &ActionResult{
		Success:   true,
		MessageID: int64(sent.MessageID),
	}, nil
}

// sendStickerAction sends a sticker
func (c *Client) sendStickerAction(action *Action) (tgbotapi.Message, error) {
	var file tgbotapi.RequestFileData
	sticker := action.Content.Attachment.Sticker
	if len(sticker) > 100 || (len(sticker) > 0 && sticker[0] == 'h') {
		file = tgbotapi.FileURL(sticker)
	} else {
		file = tgbotapi.FileID(sticker)
	}
	msg := tgbotapi.NewSticker(action.User.TgID, file)
	return c.bot.Send(msg)
}

// sendDiceAction sends a dice animation
func (c *Client) sendDiceAction(action *Action) (tgbotapi.Message, error) {
	msg := tgbotapi.NewDice(action.User.TgID)
	if action.Content.Attachment != nil && action.Content.Attachment.Dice != "" {
		msg.Emoji = action.Content.Attachment.Dice
	}
	return c.bot.Send(msg)
}

// sendContactAction sends a contact
func (c *Client) sendContactAction(action *Action) (tgbotapi.Message, error) {
	cont, ok := action.Content.Attachment.Contact.(map[string]interface{})
	if !ok {
		return tgbotapi.Message{}, nil
	}

	phoneNumber, _ := cont["phone_number"].(string)
	firstName, _ := cont["first_name"].(string)

	msg := tgbotapi.NewContact(action.User.TgID, phoneNumber, firstName)
	if lastName, ok := cont["last_name"].(string); ok {
		msg.LastName = lastName
	}
	if vcard, ok := cont["vcard"].(string); ok {
		msg.VCard = vcard
	}
	return c.bot.Send(msg)
}

// sendPollAction sends a poll
func (c *Client) sendPollAction(action *Action, parseMode string) (tgbotapi.Message, error) {
	poll, ok := action.Content.Attachment.Poll.(map[string]interface{})
	if !ok {
		return tgbotapi.Message{}, nil
	}

	question, _ := poll["question"].(string)
	var options []string
	if opts, ok := poll["options"].([]interface{}); ok {
		for _, opt := range opts {
			if s, ok := opt.(string); ok {
				options = append(options, s)
			}
		}
	}

	msg := tgbotapi.NewPoll(action.User.TgID, question, options...)

	if isAnonymous, ok := poll["is_anonymous"].(bool); ok {
		msg.IsAnonymous = isAnonymous
	}
	if pollType, ok := poll["type"].(string); ok {
		msg.Type = pollType
	}
	if allowsMultiple, ok := poll["allows_multiple_answers"].(bool); ok {
		msg.AllowsMultipleAnswers = allowsMultiple
	}
	if explanation, ok := poll["explanation"].(string); ok {
		if parseMode == "MarkdownV2" {
			explanation = FormatMarkdownV2(explanation)
		}
		msg.Explanation = explanation
		msg.ExplanationParseMode = parseMode
	}

	return c.bot.Send(msg)
}

// sendGameAction sends a game
func (c *Client) sendGameAction(action *Action) (tgbotapi.Message, error) {
	msg := tgbotapi.GameConfig{
		BaseChat:      tgbotapi.BaseChat{ChatID: action.User.TgID},
		GameShortName: action.Content.Attachment.GameShortName,
	}
	return c.bot.Send(msg)
}

// sendVenueAction sends a venue
func (c *Client) sendVenueAction(action *Action) (tgbotapi.Message, error) {
	venue, ok := action.Content.Attachment.Venue.(map[string]interface{})
	if !ok {
		return tgbotapi.Message{}, nil
	}

	latitude, _ := venue["latitude"].(float64)
	longitude, _ := venue["longitude"].(float64)
	title, _ := venue["title"].(string)
	address, _ := venue["address"].(string)

	msg := tgbotapi.NewVenue(action.User.TgID, title, address, latitude, longitude)
	if foursquareID, ok := venue["foursquare_id"].(string); ok {
		msg.FoursquareID = foursquareID
	}
	if foursquareType, ok := venue["foursquare_type"].(string); ok {
		msg.FoursquareType = foursquareType
	}
	return c.bot.Send(msg)
}

// sendTextBasedAction handles text, inline_keyboard, virtual_keyboard messages
func (c *Client) sendTextBasedAction(ctx context.Context, action *Action, text, parseMode string, callbackSaver CallbackSaver) (tgbotapi.Message, error) {
	chatID := action.User.TgID

	// Check if there's an attachment (media message)
	if action.Content.Attachment != nil && action.Content.Attachment.URL != "" {
		return c.sendMediaAction(ctx, action, text, parseMode, callbackSaver)
	}

	// Plain text message
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = parseMode

	// Apply reply markup
	if err := c.applyReplyMarkup(ctx, action, &msg.BaseChat, callbackSaver); err != nil {
		return tgbotapi.Message{}, err
	}

	return c.bot.Send(msg)
}

// sendMediaAction sends a media message with caption
func (c *Client) sendMediaAction(ctx context.Context, action *Action, caption, parseMode string, callbackSaver CallbackSaver) (tgbotapi.Message, error) {
	chatID := action.User.TgID
	attachment := action.Content.Attachment

	var baseChat tgbotapi.BaseChat
	var sent tgbotapi.Message
	var err error

	switch attachment.Type {
	case "photo":
		msg := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(attachment.URL))
		msg.Caption = caption
		msg.ParseMode = parseMode
		baseChat = msg.BaseChat
		if err := c.applyReplyMarkup(ctx, action, &msg.BaseChat, callbackSaver); err != nil {
			return tgbotapi.Message{}, err
		}
		sent, err = c.bot.Send(msg)

	case "document":
		msg := tgbotapi.NewDocument(chatID, tgbotapi.FileURL(attachment.URL))
		msg.Caption = caption
		msg.ParseMode = parseMode
		baseChat = msg.BaseChat
		if err := c.applyReplyMarkup(ctx, action, &msg.BaseChat, callbackSaver); err != nil {
			return tgbotapi.Message{}, err
		}
		sent, err = c.bot.Send(msg)

	case "video":
		msg := tgbotapi.NewVideo(chatID, tgbotapi.FileURL(attachment.URL))
		msg.Caption = caption
		msg.ParseMode = parseMode
		baseChat = msg.BaseChat
		if err := c.applyReplyMarkup(ctx, action, &msg.BaseChat, callbackSaver); err != nil {
			return tgbotapi.Message{}, err
		}
		sent, err = c.bot.Send(msg)

	case "audio":
		msg := tgbotapi.NewAudio(chatID, tgbotapi.FileURL(attachment.URL))
		msg.Caption = caption
		msg.ParseMode = parseMode
		baseChat = msg.BaseChat
		if err := c.applyReplyMarkup(ctx, action, &msg.BaseChat, callbackSaver); err != nil {
			return tgbotapi.Message{}, err
		}
		sent, err = c.bot.Send(msg)

	case "voice":
		msg := tgbotapi.NewVoice(chatID, tgbotapi.FileURL(attachment.URL))
		msg.Caption = caption
		msg.ParseMode = parseMode
		baseChat = msg.BaseChat
		if err := c.applyReplyMarkup(ctx, action, &msg.BaseChat, callbackSaver); err != nil {
			return tgbotapi.Message{}, err
		}
		sent, err = c.bot.Send(msg)

	case "video_note":
		msg := tgbotapi.NewVideoNote(chatID, 240, tgbotapi.FileURL(attachment.URL))
		baseChat = msg.BaseChat
		if err := c.applyReplyMarkup(ctx, action, &msg.BaseChat, callbackSaver); err != nil {
			return tgbotapi.Message{}, err
		}
		sent, err = c.bot.Send(msg)

	default:
		// Fallback to text message
		msg := tgbotapi.NewMessage(chatID, caption)
		msg.ParseMode = parseMode
		baseChat = msg.BaseChat
		if err := c.applyReplyMarkup(ctx, action, &msg.BaseChat, callbackSaver); err != nil {
			return tgbotapi.Message{}, err
		}
		sent, err = c.bot.Send(msg)
	}

	_ = baseChat // suppress unused variable warning
	return sent, err
}

// applyReplyMarkup applies keyboard markup to the message
func (c *Client) applyReplyMarkup(ctx context.Context, action *Action, baseChat *tgbotapi.BaseChat, callbackSaver CallbackSaver) error {
	// If custom reply_markup is provided
	if action.Content.ReplyMarkup != nil {
		markup, err := c.convertReplyMarkup(ctx, action, callbackSaver)
		if err != nil {
			return err
		}
		baseChat.ReplyMarkup = markup
		return nil
	}

	// Generate keyboard from buttons
	if len(action.Content.Buts) == 0 {
		return nil
	}

	colNum := 3
	if action.Content.ColumnNum != nil {
		colNum = *action.Content.ColumnNum
	}

	switch action.Content.Type {
	case "inline_keyboard":
		markup, err := c.buildInlineKeyboardMarkup(ctx, action, colNum, callbackSaver)
		if err != nil {
			return err
		}
		baseChat.ReplyMarkup = markup
	case "virtual_keyboard":
		baseChat.ReplyMarkup = c.buildReplyKeyboardMarkup(action, colNum)
	}

	return nil
}

// convertReplyMarkup converts custom reply_markup to tgbotapi format
func (c *Client) convertReplyMarkup(ctx context.Context, action *Action, callbackSaver CallbackSaver) (interface{}, error) {
	// Check for inline_keyboard in reply_markup
	if inlineKeyboard, ok := action.Content.ReplyMarkup["inline_keyboard"]; ok {
		rows, ok := inlineKeyboard.([]interface{})
		if !ok {
			return action.Content.ReplyMarkup, nil
		}

		var keyboard [][]tgbotapi.InlineKeyboardButton
		var callbackQueries []*CallbackData
		index := 0

		for _, row := range rows {
			rowItems, ok := row.([]interface{})
			if !ok {
				continue
			}

			var keyboardRow []tgbotapi.InlineKeyboardButton
			for _, item := range rowItems {
				btn, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				text, _ := btn["text"].(string)
				button := tgbotapi.InlineKeyboardButton{Text: text}

				// Check for URL button
				if url, ok := btn["url"].(string); ok {
					button.URL = &url
				} else {
					// Generate callback data
					hash := GenerateCallbackHash(index)
					button.CallbackData = &hash

					// Prepare callback data for saving
					data := &CallbackData{
						Project:   action.Project,
						UserID:    action.User.ID,
						QueryData: hash,
					}
					if action.Content.Actions != nil && index < len(action.Content.Actions) {
						data.Action = action.Content.Actions[index]
					}
					callbackQueries = append(callbackQueries, data)
					index++
				}

				keyboardRow = append(keyboardRow, button)
			}
			keyboard = append(keyboard, keyboardRow)
		}

		// Save callback data
		if callbackSaver != nil && len(callbackQueries) > 0 {
			if err := callbackSaver.SaveCallbackDataBatch(ctx, callbackQueries); err != nil {
				return nil, err
			}
		}

		return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: keyboard}, nil
	}

	// Check for regular keyboard
	if keyboard, ok := action.Content.ReplyMarkup["keyboard"]; ok {
		rows, ok := keyboard.([]interface{})
		if !ok {
			return action.Content.ReplyMarkup, nil
		}

		var replyKeyboard [][]tgbotapi.KeyboardButton
		for _, row := range rows {
			rowItems, ok := row.([]interface{})
			if !ok {
				continue
			}

			var keyboardRow []tgbotapi.KeyboardButton
			for _, item := range rowItems {
				switch v := item.(type) {
				case string:
					keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButton(v))
				case map[string]interface{}:
					text, _ := v["text"].(string)
					keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButton(text))
				}
			}
			replyKeyboard = append(replyKeyboard, keyboardRow)
		}

		markup := tgbotapi.NewReplyKeyboard(replyKeyboard...)
		if resize, ok := action.Content.ReplyMarkup["resize_keyboard"].(bool); ok {
			markup.ResizeKeyboard = resize
		}
		if oneTime, ok := action.Content.ReplyMarkup["one_time_keyboard"].(bool); ok {
			markup.OneTimeKeyboard = oneTime
		}

		return markup, nil
	}

	return action.Content.ReplyMarkup, nil
}

// buildInlineKeyboardMarkup builds inline keyboard from buttons
func (c *Client) buildInlineKeyboardMarkup(ctx context.Context, action *Action, colNum int, callbackSaver CallbackSaver) (tgbotapi.InlineKeyboardMarkup, error) {
	// Generate callback data hashes
	callbackData := make([]string, len(action.Content.Buts))
	var callbackQueries []*CallbackData

	for i := range action.Content.Buts {
		hash := GenerateCallbackHash(i)
		callbackData[i] = hash

		data := &CallbackData{
			Project:   action.Project,
			UserID:    action.User.ID,
			QueryData: hash,
		}
		if action.Content.Actions != nil && i < len(action.Content.Actions) {
			data.Action = action.Content.Actions[i]
		}
		callbackQueries = append(callbackQueries, data)
	}

	// Save callback data
	if callbackSaver != nil && len(callbackQueries) > 0 {
		if err := callbackSaver.SaveCallbackDataBatch(ctx, callbackQueries); err != nil {
			return tgbotapi.InlineKeyboardMarkup{}, err
		}
	}

	// Build keyboard
	rowCount := int(math.Ceil(float64(len(action.Content.Buts)) / float64(colNum)))
	keyboard := make([][]tgbotapi.InlineKeyboardButton, 0, rowCount)

	for i := 0; i < len(action.Content.Buts); i += colNum {
		var row []tgbotapi.InlineKeyboardButton
		for j := 0; j < colNum && (i+j) < len(action.Content.Buts); j++ {
			idx := i + j
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				action.Content.Buts[idx],
				callbackData[idx],
			))
		}
		keyboard = append(keyboard, row)
	}

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: keyboard}, nil
}

// buildReplyKeyboardMarkup builds reply keyboard from buttons
func (c *Client) buildReplyKeyboardMarkup(action *Action, colNum int) tgbotapi.ReplyKeyboardMarkup {
	rowCount := int(math.Ceil(float64(len(action.Content.Buts)) / float64(colNum)))
	keyboard := make([][]tgbotapi.KeyboardButton, 0, rowCount)

	for i := 0; i < len(action.Content.Buts); i += colNum {
		var row []tgbotapi.KeyboardButton
		for j := 0; j < colNum && (i+j) < len(action.Content.Buts); j++ {
			row = append(row, tgbotapi.NewKeyboardButton(action.Content.Buts[i+j]))
		}
		keyboard = append(keyboard, row)
	}

	return tgbotapi.ReplyKeyboardMarkup{
		Keyboard:        keyboard,
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
}
