---
name: mrg0773-ecosystem
description: Экосистема @mrg0773 библиотек. Используй когда нужно понять структуру проекта, зависимости между библиотеками, порядок обновления, или найти нужную библиотеку.
allowed-tools: Read, Grep, Glob, Bash
---

# Экосистема @mrg0773

## Обзор

22 TypeScript библиотеки для serverless-разработки на Yandex Cloud с YDB.

## Архитектурные правила

### SQL запросы только в YDB-слое!

SQL запросы и прямой доступ к БД разрешены **ТОЛЬКО** в:
- `ydb` - базовый клиент
- `ydb-cdc` - CDC функционал
- `ydb-toolkit` - утилиты YDB
- `ydb-apps` - репозитории приложений
- `ydb-users` - репозитории пользователей
- `ydb-bots` - репозитории ботов
- `ydb-tgferma` - репозитории TGFerma

**Все остальные библиотеки** (tgferma, tgferma-users, c2d, и т.д.) должны использовать репозитории из YDB-слоя, а не писать SQL напрямую!

### Логирование только через logger-error!

**НЕ используй console.log!** Только `@mrg0773/logger-error`:

```typescript
// ❌ console.log('message')
// ✅ log.info('message', { module: 'module-name', context })
```

### Обязательно указывать модуль!

```typescript
// Создать child logger с контекстом модуля
import { createChildLogger } from '@mrg0773/logger-error';
export const log = createChildLogger({ module: 'ydb-users' });
```

### Verbose режим в каждой библиотеке!

```typescript
// Normal: info, warn, error
// Verbose (VERBOSE=true): + debug
log.debug('Detailed info', { module: 'ydb' }); // только в verbose
```

### Нейминг: slug запрещен, slag устарел!

```typescript
// ❌ ЗАПРЕЩЕНО - slug
interface User { slug: string; }
const userSlug = getSlug();

// ⚠️ УСТАРЕЛО - slag (заменить на id)
interface App { slag: string; }  // → appId: string

// ✅ ПРАВИЛЬНО - используй id
interface User { odnoklassnikiId: string; }
interface App { appId: string; }
```

## Зависимости (порядок обновления)

```
┌─────────────────────────────────────────────────────────┐
│                     BASE LAYER                          │
│  logger-error-lib  trace-lib  schemas  types  redis     │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                      YDB LAYER                          │
│                        ydb                              │
│                         ↓                               │
│                    ydb-toolkit                          │
│                         ↓                               │
│           ydb-apps   ydb-users   ydb-bots               │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                     CORE LAYER                          │
│                    ydb-tgferma                          │
│                         ↓                               │
│                      tgferma                            │
│                         ↓                               │
│              tgferma-users   tgferma-admin              │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                   EXTERNAL LAYER                        │
│       c2d   telegram-lib   openai-lib   и другие        │
└─────────────────────────────────────────────────────────┘
```

## Все библиотеки

### Base Layer (нет внутренних зависимостей)
| Библиотека | Путь | Назначение |
|------------|------|------------|
| logger-error-lib | ~/Cursor/logger-error-lib | Логирование, форматирование ошибок |
| trace-lib | ~/Cursor/trace-lib | Distributed tracing, correlation ID |
| schemas | ~/Cursor/schemas | JSON Schema валидация |
| types | ~/Cursor/types | Общие TypeScript типы |
| redis | ~/Cursor/redis | Redis клиент, кэширование |

### YDB Layer
| Библиотека | Путь | Зависит от |
|------------|------|------------|
| ydb | ~/Cursor/ydb | base layer |
| ydb-toolkit | ~/Cursor/ydb-toolkit | ydb |
| ydb-apps | ~/Cursor/ydb-apps | ydb-toolkit |
| ydb-users | ~/Cursor/ydb-users | ydb-toolkit |
| ydb-bots | ~/Cursor/ydb-bots | ydb-toolkit |

### Core Layer (TGFerma)
| Библиотека | Путь | Зависит от |
|------------|------|------------|
| ydb-tgferma | ~/Cursor/ydb-tgferma | ydb-toolkit |
| tgferma | ~/Cursor/tgferma | ydb-tgferma |
| tgferma-users | ~/Cursor/tgferma-users | tgferma |
| tgferma-admin | ~/Cursor/tgferma-admin | tgferma |

### External Layer
| Библиотека | Путь | Назначение |
|------------|------|------------|
| yandex_cloud | ~/Cursor/yandex_cloud_lib | Yandex Cloud SDK, S3, MQ, AI |
| c2d | ~/Cursor/c2d | Click-to-Dial телефония |
| telegram-lib | ~/Cursor/telegram-lib | Telegram Bot API |
| openai-lib | ~/Cursor/openai-lib | OpenAI интеграция |

### ⛔ Legacy (НЕ ИСПОЛЬЗОВАТЬ!)
| Библиотека | Причина | Замена |
|------------|---------|--------|
| tgbots | CommonJS, устаревший код | ydb-bots, telegram-lib, redis |
| botferma20 | Legacy репозиторий, содержит tgbots | Современные core libs |

## Рабочая директория

Все библиотеки в `~/Cursor/`:

```
~/Cursor/
├── serverless-function-lib/   # Главный репозиторий
├── logger-error-lib/
├── trace-lib/
├── schemas/
├── types/
├── redis/
├── ydb/
├── ydb-toolkit/
├── ydb-apps/
├── ydb-users/
├── ydb-bots/
├── ydb-tgferma/
├── tgferma/
├── tgferma-users/
├── tgferma-admin/
├── yandex_cloud_lib/
├── c2d/
├── telegram-lib/
└── openai-lib/
```

## Технологии

- **Язык**: TypeScript (strict mode)
- **Runtime**: Node.js 22
- **База данных**: YDB (Yandex Database)
- **Кэш**: Redis
- **Линтер**: Biome
- **Ошибки**: neverthrow (не try-catch!)
- **CI/CD**: GitLab CI
- **Registry**: GitHub Packages

## Как найти библиотеку

Если нужно:
- **Логировать** → logger-error-lib
- **Трейсить запросы** → trace-lib
- **Валидировать данные** → schemas
- **Общие типы** → types
- **Кэшировать** → redis
- **Работать с YDB** → ydb, ydb-toolkit
- **TGFerma логика** → tgferma, ydb-tgferma
- **Telegram боты** → telegram-lib
- **AI/GPT** → openai-lib
- **Телефония** → c2d

## Главный репозиторий

`serverless-function-lib` - управляет всей экосистемой:
- Каскадное обновление
- CI/CD шаблоны
- Документация
- Skills для Claude Code
