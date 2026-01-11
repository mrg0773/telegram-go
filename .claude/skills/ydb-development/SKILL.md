---
name: ydb-development
description: Работа с YDB (Yandex Database) в @mrg0773. Запросы, транзакции, DDL, миграции, кэширование. Используй для работы с базой данных, SQL запросов, схем таблиц.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# YDB Development

## Быстрый старт

```typescript
import { YDBGlobalConnection } from '@mrg0773/ydb';

// Инициализация (один раз при старте)
await YDBGlobalConnection.initialize('myproject');

// Получение инстанса
const ydb = YDBGlobalConnection.getYDB();
```

## Zero-Configuration

YDBGlobalConnection автоматически определяет:
- **База данных**: по имени проекта (dev, dedic, demo, serverless-legacy)
- **Credentials**: embedded локально, IAM roles в Cloud Functions
- Не нужны .env, config файлы, service account keys!

## Пресеты и окружения (ВАЖНО!)

### Три пресета БД

| Пресет | Алиас | Назначение | Database ID |
|--------|-------|------------|-------------|
| `demo` | `demo` | **Dev/Demo** — для разработки и тестов | `etnue382v2cht2shpocg` |
| `dedicated` | `dedic` | **Production** — боевая БД | `etn3agvdb7m2itbrtrei` |
| `serverless-legacy` | `` (пустой) | Legacy serverless | — |

### Кто может выбирать пресет

**ТОЛЬКО** библиотеки `@mrg0773/ydb` и `@mrg0773/ydb-cdc` работают с пресетами напрямую.

**ВСЕ ОСТАЛЬНЫЕ** библиотеки и сервисы:
- НЕ указывают `base` при инициализации
- Получают подключение через `YDBGlobalConnection.getYDB()`
- Пресет определяется автоматически по окружению (Cloud Function получает правильную БД)

```typescript
// ❌ НЕПРАВИЛЬНО (в обычных сервисах)
await YDBGlobalConnection.initialize('myproject', { base: 'dedic' });

// ✅ ПРАВИЛЬНО (в обычных сервисах)
await YDBGlobalConnection.initialize('myproject');
// или просто
const ydb = YDBGlobalConnection.getYDB();
```

### Тесты

**КРИТИЧНО**: Тесты должны использовать `demo`, а не `dedicated`!

```typescript
// В тестах — ВСЕГДА demo (только в @mrg0773/ydb или ydb-cdc!)
beforeAll(async () => {
  await YDBGlobalConnection.initialize('test-project', { base: 'demo' });
});
```

### Как проверить к какой БД подключен

```typescript
const ydbResult = YDBGlobalConnection.getYDB();
if (ydbResult.isOk()) {
  console.log('Project:', YDBGlobalConnection.getProjectName());
  console.log('Options:', YDBGlobalConnection.getInitOptions());
}
```

## Основные операции

### SQL запросы

```typescript
// Template literal (рекомендуется)
const users = await ydb.query`
  SELECT * FROM users
  WHERE status = ${status}
  AND created_at > ${date}
`;

// С параметрами
const result = await ydb.query`
  SELECT * FROM orders
  WHERE user_id = ${userId}
  LIMIT ${limit}
`;
```

### Upsert (Insert/Update)

```typescript
// Один объект
await ydb.upsert('users', {
  id: 1,
  name: 'John',
  age: 30
});

// Несколько полей
await ydb.upsert('users', {
  id: 1,
  balance: 100,
  updated_at: new Date()
});
```

### Batch Insert

```typescript
// Массовая вставка с чанками
await ydb.batchInsert('events', largeArray, 100);
// Автоматически разбивает на батчи по 100 записей
```

### Транзакции

```typescript
const result = await ydb.transaction(async (ydb) => {
  // Все операции атомарны
  await ydb.query`UPDATE accounts SET balance = balance - ${amount} WHERE id = ${fromId}`;
  await ydb.query`UPDATE accounts SET balance = balance + ${amount} WHERE id = ${toId}`;

  return { success: true };
});
```

## DDL операции

```typescript
import { YDBDirectDDL } from '@mrg0773/ydb';

const ddl = new YDBDirectDDL(connection);

// Получить описание таблицы
const schema = await ddl.describeTable('users');

// Создать таблицу
await ddl.createTable('new_table', {
  columns: [
    { name: 'id', type: 'Uint64' },
    { name: 'name', type: 'Utf8' },
    { name: 'created_at', type: 'Timestamp' }
  ],
  primaryKey: ['id']
});

// Добавить колонку
await ddl.alterTable('users', {
  addColumns: [{ name: 'email', type: 'Utf8' }]
});
```

## Кэширование запросов

```typescript
import { YDBQueryCache } from '@mrg0773/ydb';

const cache = new YDBQueryCache(redisClient, {
  ttl: 300,  // 5 минут
  namespace: 'myapp'
});

// Запрос с кэшированием
const users = await cache.query('active_users', async () => {
  return await ydb.query`SELECT * FROM users WHERE active = true`;
});
```

## Task Queue

```typescript
import { YDBTaskQueue } from '@mrg0773/ydb';

const queue = new YDBTaskQueue(ydb, 'tasks');

// Добавить задачу
await queue.push({
  type: 'send_email',
  payload: { userId: 123, template: 'welcome' }
});

// Обработать задачи
await queue.process(async (task) => {
  // Обработка задачи
  return { success: true };
});
```

## Типы данных YDB

| YDB Type | TypeScript | Примечание |
|----------|------------|------------|
| Int32 | number | -2^31 до 2^31-1 |
| Int64 | bigint/string | -2^63 до 2^63-1 |
| Uint64 | bigint/string | 0 до 2^64-1 |
| Utf8 | string | UTF-8 строка |
| String | Buffer | Бинарные данные |
| Bool | boolean | true/false |
| Double | number | 64-bit float |
| Timestamp | Date | Микросекунды |
| Json | object | JSON документ |
| JsonDocument | object | Оптимизированный JSON |

## Обработка ошибок

```typescript
import { ok, err, Result } from '@mrg0773/logger-error';

async function getUser(id: string): Promise<Result<User, AppError>> {
  const result = await ydb.query`SELECT * FROM users WHERE id = ${id}`;

  if (result.isErr()) {
    return err(createDatabaseError('Failed to get user', {
      context: { userId: id }
    }));
  }

  const user = result.value[0];
  if (!user) {
    return err(createNotFoundError('User not found'));
  }

  return ok(user);
}
```

## Retry Logic

YDB автоматически ретраит при:
- `OVERLOADED` - перегрузка
- `UNAVAILABLE` - недоступность
- `TIMEOUT` - таймаут
- `SESSION_BUSY` - сессия занята

НЕ ретраит при:
- Синтаксические ошибки SQL
- Ошибки схемы
- Ошибки валидации

```typescript
// Кастомный retry
import { retryHelper } from '@mrg0773/ydb';

const result = await retryHelper(
  async () => await ydb.query`...`,
  {
    maxRetries: 5,
    initialDelay: 100,
    maxDelay: 5000,
    onRetry: (attempt, error) => {
      console.log(`Retry ${attempt}: ${error.message}`);
    }
  }
);
```

## Ключевые файлы

| Файл | Назначение |
|------|------------|
| `ydb/src/services/YDBGlobalConnection.ts` | Главный API |
| `ydb/src/services/YDBConnectionManager.ts` | Управление соединениями |
| `ydb/src/services/YDBDirect.ts` | Выполнение запросов |
| `ydb/src/services/ddl/YDBDirectDDL.ts` | DDL операции |
| `ydb/src/helpers/retryHelper.ts` | Логика ретраев |
| `ydb/src/helpers/sqlErrorAnalyzer.ts` | Анализ ошибок |

## Best Practices

1. **Один initialize()** в начале приложения
2. **Используй template literals** для безопасных запросов
3. **Батчи для массовых операций** (batchInsert)
4. **Транзакции для атомарности** критичных операций
5. **Кэширование** часто читаемых данных
6. **Не храни большие JSON** в колонках (используй JsonDocument)
