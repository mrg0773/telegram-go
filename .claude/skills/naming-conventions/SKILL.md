---
name: naming-conventions
description: Правила нейминга в @mrg0773. Запрещенные и устаревшие термины, правильные названия переменных и полей.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Naming Conventions @mrg0773

## Запрещенные термины

### ❌ slug - ЗАПРЕЩЕН

```typescript
// ❌ НИКОГДА не используй
interface User { slug: string; }
const userSlug = getSlug();
function generateSlug() {}

// ✅ Используй конкретные названия
interface User { odnoklassnikiId: string; }
const odnoklassnikiId = getId();
```

## Устаревшие термины

### ⚠️ slag - УСТАРЕЛ (Issue #5)

`slag` использовался как идентификатор приложения. Заменяется на:
- `appId` - для ID приложения
- `project` - для имени проекта

```typescript
// ⚠️ УСТАРЕЛО
interface App { slag: string; }
ctx.slag

// ✅ ПРАВИЛЬНО
interface App { appId: string; }
ctx.appId
// или
interface App { project: string; }
ctx.project
```

### Статистика slag (требует рефакторинга)

| Репозиторий | Вхождений |
|-------------|-----------|
| ydb-bots | 336 |
| ydb-tgferma | 317 |
| ydb-users | 247 |
| ydb-apps | 187 |
| tgferma-users | 113 |
| tgferma | 96 |
| types | 75 |
| tgferma-admin | 37 |
| ydb-toolkit | 25 |
| telegram-lib | 16 |
| schemas | 7 |
| ydb | 6 |
| c2d | 3 |

## Правильный нейминг

### Идентификаторы

```typescript
// ✅ Используй суффикс Id
userId: string;
appId: string;
botId: string;
chatId: number;
messageId: number;

// ❌ Не используй
user_id: string;  // camelCase!
userID: string;   // Id, не ID
```

### Булевы значения

```typescript
// ✅ Префикс is/has/can/should
isActive: boolean;
hasAccess: boolean;
canEdit: boolean;
shouldNotify: boolean;

// ❌ Не используй
active: boolean;
access: boolean;
```

### Коллекции

```typescript
// ✅ Множественное число
users: User[];
messages: Message[];
botIds: string[];

// ❌ Не используй
userList: User[];
userArray: User[];
```

## Проверка

```bash
# Найти slug в коде
grep -rn "slug" src/ | grep -v node_modules | grep -v ".test."

# Найти slag в коде
grep -rn "slag" src/ | grep -v node_modules | grep -v ".test."
```
