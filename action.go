package telegram

import (
	"context"
	"encoding/json"
	"math"
)

// Action represents a message action to execute
// This is the main entry point for sending messages from handler
type Action struct {
	Activity string      `json:"activity,omitempty"` // "message"
	Project  string      `json:"slag,omitempty"`     // Project slug
	User     ActionUser  `json:"user,omitempty"`     // User info (TgID required)
	Content  Content     `json:"content,omitempty"`  // Message content
	Token    string      `json:"-"`                  // Bot token (passed separately)
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
	Success   bool     `json:"success"`
	MessageID int64    `json:"message_id,omitempty"`
	Response  *Response `json:"response,omitempty"`
	Error     error    `json:"error,omitempty"`
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

// ExecuteAction executes a message action
// Returns ActionResult with message ID on success or error on failure
func (c *Client) ExecuteAction(ctx context.Context, action *Action, callbackSaver CallbackSaver) (*ActionResult, error) {
	if action.Content.Stream != "tg_direct" && action.Content.Stream != "" {
		// Only tg_direct stream is supported
		return &ActionResult{Success: false}, nil
	}

	// Build Telegram message params
	tgMessage := make(map[string]interface{})
	tgMessage["chat_id"] = action.User.TgID

	// Apply text formatting
	text := action.Content.Text
	if parseMode, ok := action.Content.Spices["parse_mode"]; ok {
		if parseMode == "MarkdownV2" {
			text = FormatMarkdownV2(text)
		}
	}

	// Copy spices to message params
	for k, v := range action.Content.Spices {
		tgMessage[k] = v
	}

	// Determine method and build message
	method := c.determineMethod(action, tgMessage, text)

	// Process keyboards
	if err := c.processKeyboards(ctx, action, tgMessage, callbackSaver); err != nil {
		return &ActionResult{Success: false, Error: err}, err
	}

	// Send chat action if configured
	if action.Content.Parameters.SendReaction != nil {
		_ = c.SendChatAction(ctx, action.User.TgID, *action.Content.Parameters.SendReaction)
	}

	// Send message
	resp, err := c.Call(ctx, method, tgMessage)
	if err != nil {
		return &ActionResult{Success: false, Response: resp, Error: err}, err
	}

	// Extract message ID from response
	var msgID int64
	if resp != nil && resp.Result != nil {
		var result map[string]interface{}
		if json.Unmarshal(resp.Result, &result) == nil {
			if id, ok := result["message_id"].(float64); ok {
				msgID = int64(id)
			}
		}
	}

	return &ActionResult{
		Success:   true,
		MessageID: msgID,
		Response:  resp,
	}, nil
}

// determineMethod determines Telegram API method based on content type
func (c *Client) determineMethod(action *Action, tgMessage map[string]interface{}, text string) string {
	method := ""

	switch action.Content.Type {
	case "sticker":
		tgMessage["sticker"] = action.Content.Attachment.Sticker
		method = "sendSticker"
	case "dice":
		tgMessage["emoji"] = action.Content.Attachment.Dice
		method = "sendDice"
	case "contact":
		if cont, ok := action.Content.Attachment.Contact.(map[string]interface{}); ok {
			tgMessage["phone_number"] = cont["phone_number"]
			tgMessage["first_name"] = cont["first_name"]
			tgMessage["last_name"] = cont["last_name"]
			tgMessage["vcard"] = cont["vcard"]
		}
		method = "sendContact"
	case "poll":
		if poll, ok := action.Content.Attachment.Poll.(map[string]interface{}); ok {
			for k, v := range poll {
				tgMessage[k] = v
			}
			// Format explanation if present
			if explanation, ok := tgMessage["explanation"].(string); ok {
				tgMessage["explanation"] = FormatMarkdownV2(explanation)
			}
		}
		method = "sendPoll"
	case "game":
		tgMessage["game_short_name"] = action.Content.Attachment.GameShortName
		method = "sendGame"
	case "venue":
		if venue, ok := action.Content.Attachment.Venue.(map[string]interface{}); ok {
			for k, v := range venue {
				tgMessage[k] = v
			}
		}
		method = "sendVenue"
	}

	// Text-based messages
	if method == "" && (action.Content.Type == "text" || action.Content.Type == "inline_keyboard" || action.Content.Type == "virtual_keyboard" || action.Content.Type == "") {
		if action.Content.Attachment == nil || action.Content.Attachment.URL == "" {
			tgMessage["text"] = text
			method = "sendMessage"
		} else {
			tgMessage["caption"] = text

			switch action.Content.Attachment.Type {
			case "photo":
				tgMessage["photo"] = action.Content.Attachment.URL
				method = "sendPhoto"
			case "document":
				tgMessage["document"] = action.Content.Attachment.URL
				method = "sendDocument"
			case "video":
				tgMessage["video"] = action.Content.Attachment.URL
				method = "sendVideo"
			case "audio":
				tgMessage["audio"] = action.Content.Attachment.URL
				method = "sendAudio"
			case "video_note":
				tgMessage["video_note"] = action.Content.Attachment.URL
				method = "sendVideoNote"
			case "voice":
				tgMessage["voice"] = action.Content.Attachment.URL
				method = "sendVoice"
			default:
				tgMessage["text"] = text
				method = "sendMessage"
			}
		}
	}

	return method
}

// processKeyboards handles inline and virtual keyboards
func (c *Client) processKeyboards(ctx context.Context, action *Action, tgMessage map[string]interface{}, callbackSaver CallbackSaver) error {
	// If reply_markup is already set, process it
	if action.Content.ReplyMarkup != nil {
		tgMessage["reply_markup"] = action.Content.ReplyMarkup

		// Process inline_keyboard callback data
		if inlineKeyboard, ok := action.Content.ReplyMarkup["inline_keyboard"]; ok {
			return c.processInlineKeyboard(ctx, action, tgMessage, inlineKeyboard, callbackSaver)
		}
		return nil
	}

	// Generate keyboard from buts
	if action.Content.Buts == nil || len(action.Content.Buts) == 0 {
		return nil
	}

	colNum := 3
	if action.Content.ColumnNum != nil {
		colNum = *action.Content.ColumnNum
	}

	switch action.Content.Type {
	case "inline_keyboard":
		return c.buildInlineKeyboard(ctx, action, tgMessage, colNum, callbackSaver)
	case "virtual_keyboard":
		c.buildVirtualKeyboard(action, tgMessage, colNum)
	}

	return nil
}

// buildInlineKeyboard builds inline keyboard and saves callback data
func (c *Client) buildInlineKeyboard(ctx context.Context, action *Action, tgMessage map[string]interface{}, colNum int, callbackSaver CallbackSaver) error {
	// Generate callback data hashes
	callbackData := make([]string, len(action.Content.Buts))
	for i := range action.Content.Buts {
		callbackData[i] = GenerateCallbackHash(i)
	}

	// Save callback data to database if saver provided
	if callbackSaver != nil {
		for i, hash := range callbackData {
			data := &CallbackData{
				Project:   action.Project,
				UserID:    action.User.ID,
				QueryData: hash,
			}

			if action.Content.Actions != nil && i < len(action.Content.Actions) {
				data.Action = action.Content.Actions[i]
			}

			if err := callbackSaver.SaveCallbackData(ctx, data); err != nil {
				return err
			}
		}
	}

	// Build keyboard
	rowCount := int(math.Ceil(float64(len(action.Content.Buts)) / float64(colNum)))
	keyboard := make([][]InlineKeyboardButton, rowCount)

	for i, k := 0, 0; i < len(action.Content.Buts); i += colNum {
		var row []InlineKeyboardButton
		for j := 0; j < colNum && (i+j) < len(action.Content.Buts); j++ {
			row = append(row, InlineKeyboardButton{
				Text:         action.Content.Buts[i+j],
				CallbackData: callbackData[i+j],
			})
		}
		keyboard[k] = row
		k++
	}

	tgMessage["reply_markup"] = map[string]interface{}{
		"inline_keyboard": keyboard,
	}

	return nil
}

// processInlineKeyboard processes existing inline keyboard and saves callback data
func (c *Client) processInlineKeyboard(ctx context.Context, action *Action, tgMessage map[string]interface{}, inlineKeyboard interface{}, callbackSaver CallbackSaver) error {
	rows, ok := inlineKeyboard.([]interface{})
	if !ok {
		return nil
	}

	index := 0
	var callbackQueries []*CallbackData

	for i, row := range rows {
		rowItems, ok := row.([]interface{})
		if !ok {
			continue
		}

		for j, item := range rowItems {
			btn, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			// Generate callback data
			hash := GenerateCallbackHash(index)
			btn["callback_data"] = hash

			// Update in place
			tgMessage["reply_markup"].(map[string]interface{})["inline_keyboard"].([]interface{})[i].([]interface{})[j] = btn

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
	}

	// Save all callback data in batch
	if callbackSaver != nil && len(callbackQueries) > 0 {
		if err := callbackSaver.SaveCallbackDataBatch(ctx, callbackQueries); err != nil {
			return err
		}
	}

	return nil
}

// buildVirtualKeyboard builds virtual (reply) keyboard
func (c *Client) buildVirtualKeyboard(action *Action, tgMessage map[string]interface{}, colNum int) {
	rowCount := int(math.Ceil(float64(len(action.Content.Buts)) / float64(colNum)))
	keyboard := make([][]string, rowCount)

	for i, k := 0, 0; i < len(action.Content.Buts); i += colNum {
		var row []string
		for j := 0; j < colNum && (i+j) < len(action.Content.Buts); j++ {
			row = append(row, action.Content.Buts[i+j])
		}
		keyboard[k] = row
		k++
	}

	tgMessage["reply_markup"] = map[string]interface{}{
		"keyboard":          keyboard,
		"resize_keyboard":   true,
		"one_time_keyboard": true,
	}
}
