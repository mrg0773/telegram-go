# telegram-go

Simple and lightweight Telegram Bot API client for Go.

## Installation

```bash
go get github.com/mrg0773/telegram-go
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    telegram "github.com/mrg0773/telegram-go"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()

    // Create client
    client := telegram.NewClient("YOUR_BOT_TOKEN", logger)

    ctx := context.Background()

    // Send a message
    msg, err := client.SendMessage(ctx, 123456789, "Hello, World!", nil)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Message sent: %d", msg.MessageID)
}
```

## Features

- Simple and clean API
- All major Telegram Bot API methods
- MarkdownV2 and HTML formatting helpers
- Automatic retry on rate limits
- Error type helpers (IsBlockedError, IsRateLimitError, etc.)
- Context support for cancellation
- Functional options for configuration

## Sending Messages

### Text Message

```go
// Simple text message
client.SendMessage(ctx, chatID, "Hello!", nil)

// With MarkdownV2 formatting
client.SendMessage(ctx, chatID, "*bold* and _italic_", map[string]interface{}{
    "parse_mode": telegram.ParseModeMarkdownV2,
})

// With reply keyboard
client.SendMessage(ctx, chatID, "Choose option:", map[string]interface{}{
    "reply_markup": telegram.InlineKeyboardMarkup{
        InlineKeyboard: [][]telegram.InlineKeyboardButton{
            {
                {Text: "Option 1", CallbackData: "opt1"},
                {Text: "Option 2", CallbackData: "opt2"},
            },
        },
    },
})
```

### Media Messages

```go
// Photo (by URL or file_id)
client.SendPhoto(ctx, chatID, "https://example.com/photo.jpg", "Caption", nil)

// Document
client.SendDocument(ctx, chatID, "file_id_here", "Document caption", nil)

// Video
client.SendVideo(ctx, chatID, "video_file_id", "Video caption", nil)

// Audio
client.SendAudio(ctx, chatID, "audio_file_id", "Audio caption", nil)

// Voice
client.SendVoice(ctx, chatID, "voice_file_id", "Voice caption", nil)

// Sticker
client.SendSticker(ctx, chatID, "sticker_file_id", nil)

// Location
client.SendLocation(ctx, chatID, 55.7558, 37.6173, nil) // Moscow
```

### Other Methods

```go
// Edit message
client.EditMessageText(ctx, chatID, messageID, "New text", nil)

// Delete message
client.DeleteMessage(ctx, chatID, messageID)

// Answer callback query
client.AnswerCallbackQuery(ctx, callbackQueryID, map[string]interface{}{
    "text": "Button pressed!",
})

// Send typing indicator
client.SendChatAction(ctx, chatID, "typing")

// Get file info
file, _ := client.GetFile(ctx, fileID)
downloadURL := client.GetFileURL(file.FilePath)
```

## Formatting Helpers

### MarkdownV2

```go
// Escape special characters
text := telegram.EscapeMarkdownV2("Hello! How are you?")

// Format helpers that auto-escape
bold := telegram.BoldV2("Important")
italic := telegram.ItalicV2("emphasized")
link := telegram.LinkV2("Click here", "https://example.com")
mention := telegram.MentionV2("User", 123456789)
```

### HTML

```go
text := telegram.BoldHTML("Important")
link := telegram.LinkHTML("Click here", "https://example.com")
code := telegram.CodeHTML("fmt.Println()")
```

## Error Handling

```go
msg, err := client.SendMessage(ctx, chatID, "Hello", nil)
if err != nil {
    if telegram.IsBlockedError(err) {
        // User blocked the bot
        log.Println("User blocked bot")
    } else if telegram.IsRateLimitError(err) {
        // Rate limited (auto-retry is built-in)
        log.Println("Rate limited")
    } else if telegram.IsBadRequestError(err) {
        // Invalid request
        log.Printf("Bad request: %v", err)
    } else {
        log.Printf("Error: %v", err)
    }
    return
}
```

## Configuration Options

```go
// Custom timeout
client := telegram.NewClient(token, logger,
    telegram.WithTimeout(60 * time.Second),
)

// Custom HTTP client
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns: 100,
    },
}
client := telegram.NewClient(token, logger,
    telegram.WithHTTPClient(httpClient),
)

// Custom base URL (for testing)
client := telegram.NewClient(token, logger,
    telegram.WithBaseURL("http://localhost:8081/bot"),
)
```

## Webhook Handling

```go
func webhookHandler(w http.ResponseWriter, r *http.Request) {
    var update telegram.Update
    if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
        http.Error(w, "Bad request", 400)
        return
    }

    if update.Message != nil {
        // Handle message
        client.SendMessage(ctx, update.Message.Chat.ID, "Got your message!", nil)
    }

    if update.CallbackQuery != nil {
        // Handle callback
        client.AnswerCallbackQuery(ctx, update.CallbackQuery.ID, nil)
    }

    w.WriteHeader(200)
}
```

## Action Execution (for handler integration)

The library provides `ExecuteAction` method for executing message actions from handler-go-v3.

### Action Structure

```go
type Action struct {
    Activity string      // "message"
    Project  string      // Project slug
    User     ActionUser  // User info (TgID required)
    Content  Content     // Message content
    Token    string      // Bot token
}

type ActionUser struct {
    TgID int64  // Telegram user ID (chat_id)
    ID   string // Internal user ID
}

type Content struct {
    Type        string                 // text, inline_keyboard, virtual_keyboard, sticker, dice, etc.
    Stream      string                 // "tg_direct"
    Text        string                 // Message text
    Attachment  *Attachment            // Media attachment
    Buts        []string               // Button labels
    Actions     []json.RawMessage      // Button callback actions
    ReplyMarkup map[string]interface{} // Custom reply markup
    ColumnNum   *int                   // Keyboard column count (default: 3)
    Spices      map[string]interface{} // Extra params (parse_mode, etc.)
    Parameters  Parameters             // Action parameters
}
```

### Supported Content Types

- `text` - Plain text message
- `inline_keyboard` - Text with inline keyboard
- `virtual_keyboard` - Text with reply keyboard
- `sticker` - Sticker message
- `dice` - Dice animation
- `contact` - Contact message
- `poll` - Poll message
- `game` - Game message
- `venue` - Venue message

### Attachment Types

- `photo` - Photo by URL or file_id
- `document` - Document file
- `video` - Video file
- `audio` - Audio file
- `voice` - Voice message
- `video_note` - Round video

### Example Usage

```go
// Create client
client := telegram.NewClient(botToken, logger)

// Create action
action := &telegram.Action{
    Activity: "message",
    Project:  "myproject",
    User: telegram.ActionUser{
        TgID: 123456789,
        ID:   "user123",
    },
    Content: telegram.Content{
        Type:   "inline_keyboard",
        Stream: "tg_direct",
        Text:   "Choose an option:",
        Buts:   []string{"Option 1", "Option 2", "Option 3"},
        Spices: map[string]interface{}{
            "parse_mode": "MarkdownV2",
        },
    },
}

// Execute action (with callback saver for inline keyboards)
result, err := client.ExecuteAction(ctx, action, myCallbackSaver)
if err != nil {
    log.Printf("Error: %v", err)
    return
}

log.Printf("Message sent, ID: %d", result.MessageID)
```

### Callback Data Saver Interface

For inline keyboards, implement `CallbackSaver` to store callback data:

```go
type CallbackSaver interface {
    SaveCallbackData(ctx context.Context, data *CallbackData) error
    SaveCallbackDataBatch(ctx context.Context, data []*CallbackData) error
}

type CallbackData struct {
    Project   string
    UserID    string
    QueryData string          // Generated hash
    Action    json.RawMessage // Action to execute on callback
}
```

### Smart MarkdownV2 Formatting

The `FormatMarkdownV2` function automatically escapes special characters while preserving markdown formatting:

```go
// Input: "Hello! *bold* and _italic_ with [link](https://example.com)"
// Output: "Hello\\! *bold* and _italic_ with [link](https://example.com)"

text := telegram.FormatMarkdownV2("Hello! Check *this* out.")
client.SendMessage(ctx, chatID, text, map[string]interface{}{
    "parse_mode": "MarkdownV2",
})
```

Supported formatting:
- `*bold*`
- `_italic_`
- `__underline__`
- `~strikethrough~`
- `||spoiler||`
- `` `code` ``
- ` ```code block``` `
- `[text](url)`

## License

MIT
