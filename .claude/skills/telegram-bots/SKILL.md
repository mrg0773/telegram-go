---
name: telegram-bots
description: Разработка Telegram ботов в @mrg0773. Bot API, handlers, keyboards, messages. Используй для создания ботов, отправки сообщений, обработки команд.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Telegram Bots Development

## Инициализация бота

```typescript
import { TgFermaBotService } from '@mrg0773/telegram';

// Создать или получить инстанс
const botResult = await TgFermaBotService.getOrCreateInstance({
  token: 'BOT_TOKEN',
  botId: 123456789
});

if (botResult.isErr()) {
  console.error('Failed to create bot:', botResult.error);
  return;
}

const bot = botResult.value;
```

## Bot Manager

```typescript
import { BotManager } from '@mrg0773/telegram';

const manager = new BotManager();

// Инициализация с валидацией
const result = manager.initialize({
  token: 'BOT_TOKEN',
  username: 'mybot',
  botId: 123456789
});

if (result.isErr()) {
  console.error('Validation failed:', result.error);
}
```

## Отправка сообщений

### Текстовое сообщение

```typescript
const result = await bot.sendMessage(chatId, 'Hello!', token, {
  parse_mode: 'HTML',
  disable_notification: false
});

if (result.isOk()) {
  console.log('Message sent:', result.value.message_id);
}
```

### С форматированием

```typescript
// HTML
await bot.sendMessage(chatId, `
  <b>Bold</b>
  <i>Italic</i>
  <code>Code</code>
  <pre>Preformatted</pre>
  <a href="https://example.com">Link</a>
`, token, { parse_mode: 'HTML' });

// Markdown
await bot.sendMessage(chatId, `
  *Bold*
  _Italic_
  \`Code\`
  \`\`\`
  Preformatted
  \`\`\`
  [Link](https://example.com)
`, token, { parse_mode: 'MarkdownV2' });
```

### Редактирование

```typescript
await bot.editMessageText(chatId, messageId, 'Updated text', token, {
  parse_mode: 'HTML'
});
```

### Удаление

```typescript
await bot.deleteMessage(chatId, messageId, token);
```

## Клавиатуры

### Inline Keyboard

```typescript
const keyboard = {
  inline_keyboard: [
    [
      { text: 'Button 1', callback_data: 'action_1' },
      { text: 'Button 2', callback_data: 'action_2' }
    ],
    [
      { text: 'URL Button', url: 'https://example.com' }
    ]
  ]
};

await bot.sendMessage(chatId, 'Choose:', token, {
  reply_markup: keyboard
});
```

### Reply Keyboard

```typescript
const keyboard = {
  keyboard: [
    [{ text: 'Option 1' }, { text: 'Option 2' }],
    [{ text: 'Option 3' }]
  ],
  resize_keyboard: true,
  one_time_keyboard: true
};

await bot.sendMessage(chatId, 'Choose:', token, {
  reply_markup: keyboard
});
```

### Remove Keyboard

```typescript
await bot.sendMessage(chatId, 'Keyboard removed', token, {
  reply_markup: { remove_keyboard: true }
});
```

## Медиа

### Фото

```typescript
await bot.sendPhoto(chatId, photoUrl, token, {
  caption: 'Photo caption',
  parse_mode: 'HTML'
});
```

### Документ

```typescript
await bot.sendDocument(chatId, documentUrl, token, {
  caption: 'Document description'
});
```

### Видео

```typescript
await bot.sendVideo(chatId, videoUrl, token, {
  caption: 'Video caption',
  duration: 120,
  width: 1920,
  height: 1080
});
```

### Media Group

```typescript
await bot.sendMediaGroup(chatId, [
  { type: 'photo', media: 'url1', caption: 'First' },
  { type: 'photo', media: 'url2' },
  { type: 'photo', media: 'url3' }
], token);
```

## Callback Queries

```typescript
// Обработка callback
async function handleCallback(update: Update) {
  const callback = update.callback_query;
  if (!callback) return;

  const { data, message } = callback;

  // Ответить на callback (убрать loading)
  await bot.answerCallbackQuery(callback.id, token, {
    text: 'Processing...',
    show_alert: false
  });

  // Обработать действие
  switch (data) {
    case 'action_1':
      await handleAction1(message);
      break;
    case 'action_2':
      await handleAction2(message);
      break;
  }
}
```

## Webhook Handler

```typescript
import { handleTelegramWebhook } from '@mrg0773/telegram';

export async function handler(event: APIGatewayEvent) {
  const update = JSON.parse(event.body);

  const result = await handleTelegramWebhook(update, {
    onMessage: async (message) => {
      // Обработка сообщения
    },
    onCallback: async (callback) => {
      // Обработка callback
    },
    onInlineQuery: async (query) => {
      // Обработка inline query
    }
  });

  return { statusCode: 200, body: 'ok' };
}
```

## Валидация

```typescript
import { validators } from '@mrg0773/telegram';

// Валидация токена
if (!validators.isValidToken(token)) {
  return err(createValidationError('Invalid bot token'));
}

// Валидация chat ID
if (!validators.isValidChatId(chatId)) {
  return err(createValidationError('Invalid chat ID'));
}

// Валидация сообщения
const validation = validators.validateMessage(message);
if (validation.isErr()) {
  return validation;
}
```

## Обработка ошибок

```typescript
async function sendSafe(chatId: number, text: string): Promise<Result<Message, AppError>> {
  const result = await bot.sendMessage(chatId, text, token);

  if (result.isErr()) {
    // Анализ ошибки Telegram
    const error = result.error;

    if (error.message.includes('chat not found')) {
      return errWithLog('NOT_FOUND_ERROR', 'Chat not found', {
        context: { chatId }
      });
    }

    if (error.message.includes('bot was blocked')) {
      return errWithLog('AUTH_ERROR', 'Bot was blocked by user', {
        context: { chatId }
      });
    }

    return errWithLog('API_ERROR', 'Telegram API error', {
      context: { chatId, error: error.message }
    });
  }

  return result;
}
```

## Типы

```typescript
import type {
  Update,
  Message,
  CallbackQuery,
  InlineKeyboardMarkup,
  ReplyKeyboardMarkup,
  BotInfo
} from '@mrg0773/telegram';

// Message типы
interface Message {
  message_id: number;
  chat: Chat;
  from?: User;
  text?: string;
  photo?: PhotoSize[];
  document?: Document;
  // ...
}

// Callback типы
interface CallbackQuery {
  id: string;
  from: User;
  message?: Message;
  data?: string;
}
```

## Ключевые файлы

| Файл | Назначение |
|------|------------|
| `telegram-lib/src/business/bot/bot.manager.ts` | Управление ботами |
| `telegram-lib/src/methods/messaging.ts` | Отправка сообщений |
| `telegram-lib/src/methods/media.ts` | Работа с медиа |
| `telegram-lib/src/core/validators.ts` | Валидация |
| `telegram-lib/src/contracts/bot.contracts.ts` | Интерфейсы |

## Best Practices

1. **Валидируй входные данные** перед отправкой
2. **Обрабатывай все ошибки** Telegram API
3. **Используй parse_mode** для форматирования
4. **Отвечай на callback_query** быстро (< 10 сек)
5. **Батчинг** для массовой рассылки
6. **Rate limits** - не более 30 msg/sec per chat
7. **Логируй** все взаимодействия для отладки
