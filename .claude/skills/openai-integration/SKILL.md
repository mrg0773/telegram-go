---
name: openai-integration
description: Интеграция с OpenAI API в @mrg0773. GPT, embeddings, assistants, prompts. Используй для работы с AI, генерации текста, embeddings.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# OpenAI Integration

## Клиент

```typescript
import { OpenAIClient } from '@mrg0773/openai-lib';

const client = new OpenAIClient({
  apiKey: process.env.OPENAI_API_KEY,
  organization: process.env.OPENAI_ORG_ID // опционально
});
```

## Chat Completions

### Базовый запрос

```typescript
const result = await client.chat({
  model: 'gpt-4-turbo-preview',
  messages: [
    { role: 'system', content: 'You are a helpful assistant.' },
    { role: 'user', content: 'Hello!' }
  ]
});

if (result.isOk()) {
  console.log(result.value.choices[0].message.content);
}
```

### С параметрами

```typescript
const result = await client.chat({
  model: 'gpt-4-turbo-preview',
  messages: [...],
  temperature: 0.7,      // Креативность (0-2)
  max_tokens: 1000,      // Макс. токенов в ответе
  top_p: 1,              // Nucleus sampling
  frequency_penalty: 0,  // Штраф за повторения
  presence_penalty: 0,   // Штраф за новые темы
  stop: ['\n\n']        // Стоп-последовательности
});
```

### Streaming

```typescript
const stream = await client.chatStream({
  model: 'gpt-4-turbo-preview',
  messages: [...],
  stream: true
});

for await (const chunk of stream) {
  const content = chunk.choices[0]?.delta?.content || '';
  process.stdout.write(content);
}
```

## Function Calling (Tools)

```typescript
const result = await client.chat({
  model: 'gpt-4-turbo-preview',
  messages: [
    { role: 'user', content: 'What is the weather in Moscow?' }
  ],
  tools: [
    {
      type: 'function',
      function: {
        name: 'get_weather',
        description: 'Get current weather for a location',
        parameters: {
          type: 'object',
          properties: {
            location: {
              type: 'string',
              description: 'City name'
            }
          },
          required: ['location']
        }
      }
    }
  ],
  tool_choice: 'auto'
});

// Обработка вызова функции
const toolCall = result.value.choices[0].message.tool_calls?.[0];
if (toolCall) {
  const args = JSON.parse(toolCall.function.arguments);
  const weather = await getWeather(args.location);

  // Продолжить с результатом
  const finalResult = await client.chat({
    model: 'gpt-4-turbo-preview',
    messages: [
      ...messages,
      result.value.choices[0].message,
      {
        role: 'tool',
        tool_call_id: toolCall.id,
        content: JSON.stringify(weather)
      }
    ]
  });
}
```

## Embeddings

```typescript
const result = await client.embeddings({
  model: 'text-embedding-3-small',
  input: 'Hello world'
});

if (result.isOk()) {
  const vector = result.value.data[0].embedding;
  // vector: number[] - 1536 dimensions
}

// Batch embeddings
const result = await client.embeddings({
  model: 'text-embedding-3-small',
  input: ['Text 1', 'Text 2', 'Text 3']
});
```

### Similarity Search

```typescript
function cosineSimilarity(a: number[], b: number[]): number {
  let dotProduct = 0;
  let normA = 0;
  let normB = 0;

  for (let i = 0; i < a.length; i++) {
    dotProduct += a[i] * b[i];
    normA += a[i] * a[i];
    normB += b[i] * b[i];
  }

  return dotProduct / (Math.sqrt(normA) * Math.sqrt(normB));
}

// Найти похожие документы
const queryEmbedding = await getEmbedding(query);
const similarities = documents.map(doc => ({
  doc,
  score: cosineSimilarity(queryEmbedding, doc.embedding)
}));
similarities.sort((a, b) => b.score - a.score);
```

## Assistants API

```typescript
// Создать ассистента
const assistant = await client.createAssistant({
  model: 'gpt-4-turbo-preview',
  name: 'My Assistant',
  instructions: 'You are a helpful assistant.',
  tools: [{ type: 'code_interpreter' }]
});

// Создать thread
const thread = await client.createThread();

// Добавить сообщение
await client.addMessage(thread.id, {
  role: 'user',
  content: 'Analyze this data...'
});

// Запустить
const run = await client.createRun(thread.id, {
  assistant_id: assistant.id
});

// Ждать завершения
const completed = await client.waitForRun(thread.id, run.id);

// Получить ответ
const messages = await client.listMessages(thread.id);
```

## Промпт-инжиниринг

### System Prompt

```typescript
const systemPrompt = `You are a customer support agent for TGFerma game.

Rules:
1. Be friendly and helpful
2. Answer only about the game
3. If unsure, say you don't know
4. Keep responses concise

Game info:
- TGFerma is a Telegram farming game
- Players can grow crops, raise animals
- Currency: coins and gems
`;
```

### Few-shot примеры

```typescript
const messages = [
  { role: 'system', content: systemPrompt },
  { role: 'user', content: 'How do I plant crops?' },
  { role: 'assistant', content: 'To plant crops, go to your farm and tap on an empty plot...' },
  { role: 'user', content: 'What about animals?' },
  { role: 'assistant', content: 'Animals can be purchased in the shop...' },
  { role: 'user', content: actualUserQuestion }
];
```

### Chain of Thought

```typescript
const prompt = `
Analyze the user's question step by step:

1. Identify the main topic
2. Determine what information is needed
3. Formulate a clear answer

Question: ${userQuestion}

Step 1: The main topic is...
Step 2: The user needs to know...
Step 3: Answer:
`;
```

## Обработка ошибок

```typescript
import { ok, err, Result } from '@mrg0773/logger-error';

async function generateResponse(prompt: string): Promise<Result<string, AppError>> {
  const result = await client.chat({
    model: 'gpt-4-turbo-preview',
    messages: [{ role: 'user', content: prompt }]
  });

  if (result.isErr()) {
    const error = result.error;

    // Rate limit
    if (error.message.includes('rate_limit')) {
      return errWithLog('API_ERROR', 'OpenAI rate limit exceeded', {
        context: { retryAfter: error.headers?.['retry-after'] }
      });
    }

    // Token limit
    if (error.message.includes('context_length')) {
      return errWithLog('VALIDATION_ERROR', 'Prompt too long', {
        context: { maxTokens: 128000 }
      });
    }

    return errWithLog('API_ERROR', 'OpenAI API error', {
      context: { error: error.message }
    });
  }

  return ok(result.value.choices[0].message.content || '');
}
```

## Модели

| Модель | Контекст | Стоимость | Применение |
|--------|----------|-----------|------------|
| gpt-4-turbo-preview | 128K | $$$ | Сложные задачи |
| gpt-4 | 8K | $$$ | Точность |
| gpt-3.5-turbo | 16K | $ | Простые задачи |
| text-embedding-3-small | - | ¢ | Embeddings |
| text-embedding-3-large | - | ¢¢ | Точные embeddings |

## Ключевые файлы

| Файл | Назначение |
|------|------------|
| `openai-lib/src/client.ts` | OpenAI клиент |
| `openai-lib/src/prompts/` | Шаблоны промптов |
| `openai-lib/src/embeddings/` | Работа с embeddings |

## Best Practices

1. **System prompt** - чёткие инструкции
2. **Temperature** - низкая для фактов, высокая для креатива
3. **Max tokens** - ограничивай для экономии
4. **Retry** - при rate limit с exponential backoff
5. **Кэширование** - embeddings в Redis/YDB
6. **Мониторинг** - отслеживай токены и стоимость
7. **Fallback** - запасная модель при ошибках
