---
name: tgferma-game
description: Бизнес-логика TGFerma. Игровая механика, управление ресурсами, боты, пользователи. Используй для работы с TGFerma, фермами, игровыми сущностями.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# TGFerma Game Logic

## Архитектура

```
tgferma/
├── context/       # Global context management
├── methods/       # Business operations
├── services/      # Domain services
├── contracts/     # Interface definitions
├── types/         # Type definitions
└── helpers/       # Error handling, logging
```

## Инициализация

```typescript
import { BotContextManager } from '@mrg0773/tgferma';

// Один раз при старте
const initResult = await BotContextManager.initialize('myproject');

if (initResult.isErr()) {
  console.error('Failed to initialize:', initResult.error);
  return;
}

// Получение контекста
const context = BotContextManager.getInstance();
```

## Основные методы

### Получение данных бота

```typescript
import { getBot, listBots, getBotStats } from '@mrg0773/tgferma';

// Один бот
const botResult = await getBot(project, botId);
if (botResult.isOk()) {
  const bot = botResult.value;
  console.log(bot.username, bot.status);
}

// Все боты проекта
const botsResult = await listBots(project);

// Статистика
const stats = await BotContextManager.getBotStats();
```

### Работа с сообщениями

```typescript
import { getInboxMessages, sendMessage, markAsRead } from '@mrg0773/tgferma';

// Входящие сообщения
const messages = await getInboxMessages(project, userId, {
  limit: 50,
  offset: 0
});

// Отправка
const result = await sendMessage(project, {
  chatId: userId,
  text: 'Hello!',
  parseMode: 'HTML'
});

// Отметить прочитанным
await markAsRead(project, messageId);
```

### Управление настройками

```typescript
import { getSaveSetup, updateSetup, getSetupVersions } from '@mrg0773/tgferma';

// Получить настройки
const setup = await getSaveSetup(project, setupId);

// Обновить
const result = await updateSetup(project, setupId, {
  name: 'New name',
  config: { /* ... */ }
});

// История версий
const versions = await getSetupVersions(project, setupId);
```

### Магазин и продукты

```typescript
import { getProducts, createProduct, updateProduct } from '@mrg0773/tgferma';

// Список продуктов
const products = await getProducts(project);

// Создать
const result = await createProduct(project, {
  name: 'New Product',
  price: 100,
  description: 'Product description'
});

// Обновить
await updateProduct(project, productId, {
  price: 150
});
```

## safeRepoCall

Безопасный вызов репозитория с логированием:

```typescript
import { safeRepoCall } from '@mrg0773/tgferma';

async function getUser(id: string) {
  return safeRepoCall(
    () => userRepo.findById(id),
    'getUser',  // Имя операции для логов
    { userId: id }  // Контекст
  );
}
// Автоматически:
// - Ловит ошибки
// - Логирует с контекстом
// - Возвращает Result<T, TgFermaError>
```

## Типы и контракты

```typescript
import type {
  BotDataWithStatus,
  SetupInfo,
  SetupVersion,
  UserAction,
  MessageSendOptions,
  ProductInfo
} from '@mrg0773/tgferma';

interface BotDataWithStatus {
  id: number;
  username: string;
  token: string;
  status: 'active' | 'inactive' | 'banned';
  createdAt: Date;
  config: BotConfig;
}

interface SetupInfo {
  id: string;
  name: string;
  version: number;
  config: Record<string, unknown>;
  updatedAt: Date;
}
```

## Работа с YDB репозиториями

```typescript
import { ydbTgfermaRepos } from '@mrg0773/ydb-tgferma';

// Получить репозитории
const { userRepo, botRepo, messageRepo, setupRepo } = ydbTgfermaRepos;

// Использование
const users = await userRepo.findByProject(projectId);
const bots = await botRepo.findActive();
```

## Игровая механика

### Ресурсы

```typescript
import { ResourceManager } from '@mrg0773/tgferma';

const manager = new ResourceManager(context);

// Получить баланс
const balance = await manager.getBalance(userId);

// Начислить ресурсы
const result = await manager.addResources(userId, {
  coins: 100,
  gems: 10
});

// Списать
const result = await manager.spendResources(userId, {
  coins: 50
});
```

### Действия пользователя

```typescript
import { trackUserAction } from '@mrg0773/tgferma';

// Отслеживание действий
await trackUserAction(project, {
  userId,
  action: 'purchase',
  data: {
    productId: '123',
    amount: 100
  }
});
```

## Middleware

```typescript
import { authMiddleware, rateLimitMiddleware } from '@mrg0773/tgferma';

// Аутентификация
const authResult = await authMiddleware(request);
if (authResult.isErr()) {
  return unauthorized();
}

// Rate limiting
const rateResult = await rateLimitMiddleware(userId, {
  maxRequests: 100,
  windowMs: 60000
});
if (rateResult.isErr()) {
  return tooManyRequests();
}
```

## Обработка ошибок

```typescript
import { TgFermaError, handleTgFermaError } from '@mrg0773/tgferma';

async function processAction(action: UserAction): Promise<Result<void, TgFermaError>> {
  // Валидация
  if (!action.userId) {
    return errWithLog('VALIDATION_ERROR', 'User ID required');
  }

  // Бизнес-логика
  const result = await executeAction(action);

  if (result.isErr()) {
    return handleTgFermaError(result.error, {
      action: action.type,
      userId: action.userId
    });
  }

  return ok(undefined);
}
```

## Ключевые файлы

| Путь | Назначение |
|------|------------|
| `tgferma/src/context/` | Singleton context |
| `tgferma/src/methods/` | Business operations |
| `tgferma/src/services/` | Domain services |
| `ydb-tgferma/src/repositories/` | Data access |
| `ydb-tgferma/src/models/` | Data models |

## Зависимости

```
@mrg0773/tgferma
    └── @mrg0773/ydb-tgferma
        └── @mrg0773/ydb-toolkit
            └── @mrg0773/ydb
                └── @mrg0773/logger-error
```

## Best Practices

1. **Используй BotContextManager** для инициализации
2. **safeRepoCall** для всех операций с БД
3. **Result типы** для обработки ошибок
4. **Валидируй** входные данные
5. **Логируй** все важные операции
6. **Транзакции** для атомарных операций
