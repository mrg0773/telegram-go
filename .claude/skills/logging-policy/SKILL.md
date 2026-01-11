---
name: logging-policy
description: Политика логирования в @mrg0773. Используй @mrg0773/logger-error, не console.log. Трейсинг через @mrg0773/trace-lib.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Политика логирования @mrg0773

## Правило #1: НЕ используй console.log!

```typescript
// ❌ ПЛОХО
console.log('User created:', userId);
console.error('Failed to save:', error);
console.warn('Deprecated method');

// ✅ ХОРОШО
import { log } from '@mrg0773/logger-error';

log.info('User created', { userId });
log.error('Failed to save', { error: error.message, stack: error.stack });
log.warn('Deprecated method', { method: 'oldMethod' });
```

## Почему?

1. **Структурированные логи** - JSON формат для CloudWatch
2. **Единый формат** - легко искать и фильтровать
3. **Контекст** - trace ID, timestamp, уровень
4. **Production-ready** - правильный output для Yandex Cloud

## Уровни логирования

```typescript
import { log } from '@mrg0773/logger-error';

log.debug('Detailed info');  // Только в verbose режиме
log.info('Normal operation'); // Стандартные события
log.warn('Something suspicious'); // Предупреждения
log.error('Something failed'); // Ошибки
```

## Идентификация модуля

**ОБЯЗАТЕЛЬНО** указывать имя модуля в каждом логе:

```typescript
import { log } from '@mrg0773/logger-error';

// ❌ ПЛОХО - непонятно откуда лог
log.info('User created', { userId });

// ✅ ХОРОШО - указан модуль
log.info('User created', { module: 'ydb-users', userId });

// ✅ ЕЩЕ ЛУЧШЕ - создать child logger с контекстом
import { createChildLogger } from '@mrg0773/logger-error';
const log = createChildLogger({ module: 'ydb-users' });
log.info('User created', { userId });
```

## Verbose режим

Каждая библиотека должна поддерживать два режима:

### Normal режим (по умолчанию)
- Логирует: `info`, `warn`, `error`
- Минимум информации для production

### Verbose режим
- Логирует: `debug`, `info`, `warn`, `error`
- Детальная информация для отладки
- Включается через `VERBOSE=true` или `LOG_LEVEL=debug`

```typescript
import { log } from '@mrg0773/logger-error';

const isVerbose = process.env.VERBOSE === 'true' || process.env.LOG_LEVEL === 'debug';

// Verbose-only логи
if (isVerbose) {
  log.debug('SQL query', { module: 'ydb', query, params });
}

// Или использовать уровень debug (фильтруется автоматически)
log.debug('Detailed operation', { module: 'ydb-toolkit', data });
```

### Реализация в библиотеке

```typescript
// lib/logger.ts
import { createChildLogger } from '@mrg0773/logger-error';

const MODULE_NAME = 'ydb-users';

export const log = createChildLogger({ module: MODULE_NAME });

export const isVerbose = () =>
  process.env.VERBOSE === 'true' ||
  process.env.LOG_LEVEL === 'debug';
```

## Структура лога

```typescript
log.info('User created', {
  userId: '123',
  action: 'create',
  duration: 150
});

// Output (JSON):
{
  "level": "info",
  "message": "User created",
  "timestamp": "2024-01-15T10:30:00.000Z",
  "context": {
    "userId": "123",
    "action": "create",
    "duration": 150
  }
}
```

## Трейсинг

```typescript
import { TraceContext, generateTraceId } from '@mrg0773/trace-lib';

// Создать trace
const traceId = generateTraceId();
const context = new TraceContext(traceId);

// Передать в логи
log.info('Processing request', { traceId });
```

## Текущие нарушения

| Репозиторий | console calls | Issue |
|-------------|---------------|-------|
| telegram-lib | 275 | #2 |
| c2d | 250 | #6 |
| redis | 152 | #2 |
| ydb | 86 | #3 |
| tgferma | 59 | #56 |

## Проверка

```bash
# Найти console.log в production коде
grep -rn "console\.\(log\|error\|warn\)" src/ | grep -v ".test." | grep -v ".spec."
```

## Миграция

```typescript
// Шаг 1: Добавить импорт
import { log } from '@mrg0773/logger-error';

// Шаг 2: Найти и заменить
// console.log(msg) → log.info(msg)
// console.error(msg) → log.error(msg)
// console.warn(msg) → log.warn(msg)
// console.info(msg) → log.info(msg)

// Шаг 3: Добавить контекст
// console.log('User:', user) → log.info('User', { user })
```

## Исключения

Можно использовать console в:
- CLI инструментах (`cli.ts`)
- Тестах (`*.test.ts`, `*.spec.ts`)
- JSDoc примерах (в комментариях)

## Best Practices

1. **Всегда добавляй контекст** - что, где, когда
2. **Не логируй sensitive data** - пароли, токены
3. **Используй правильный уровень** - info для нормы, error для ошибок
4. **Добавляй traceId** - для distributed tracing
5. **Структурируй данные** - объекты вместо строк
