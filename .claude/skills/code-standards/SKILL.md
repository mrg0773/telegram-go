---
name: code-standards
description: Стандарты кода @mrg0773. Конвенции именования, PR, code review, git workflow. Используй для code review, стандартов, best practices.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Code Standards

## Именование

### Файлы

```
kebab-case.ts          # Обычные файлы
PascalCase.ts          # React компоненты (если используются)
kebab-case.test.ts     # Тесты
kebab-case.spec.ts     # Альтернатива для тестов
```

### Переменные и функции

```typescript
// camelCase для переменных и функций
const userName = 'John';
const isActive = true;

function getUserById(id: string) { }
async function fetchUserData() { }

// SCREAMING_SNAKE_CASE для констант
const MAX_RETRY_COUNT = 3;
const DEFAULT_TIMEOUT_MS = 5000;
const API_ENDPOINTS = { ... };
```

### Классы и типы

```typescript
// PascalCase для классов и типов
class UserService { }
interface UserData { }
type UserId = string;
enum UserStatus { Active, Inactive }
```

### Приватные поля

```typescript
class Service {
  // Используем # для приватных полей (ES2022)
  #privateField: string;

  // Или _ префикс для совместимости
  private _legacyPrivate: string;
}
```

## Структура файла

```typescript
// 1. Импорты (сортировка biome)
import { ok, err, Result } from '@mrg0773/logger-error';
import { YDBGlobalConnection } from '@mrg0773/ydb';

import { UserRepository } from './repositories/user.repository';
import { validateUser } from './validators';

import type { User, CreateUserDto } from './types';

// 2. Константы
const DEFAULT_LIMIT = 10;

// 3. Типы (если локальные)
interface ServiceConfig {
  maxRetries: number;
}

// 4. Основной код
export class UserService {
  // ...
}

// 5. Хелперы (в конце файла или отдельном файле)
function formatUserName(name: string): string {
  return name.trim().toLowerCase();
}
```

## Функции

### Чистые функции

```typescript
// ✅ Хорошо - чистая функция
function calculateTotal(items: Item[]): number {
  return items.reduce((sum, item) => sum + item.price, 0);
}

// ❌ Плохо - side effects
function calculateTotal(items: Item[]): number {
  console.log('Calculating...'); // side effect
  globalTotal = items.reduce(...); // mutation
  return globalTotal;
}
```

### Размер функции

```typescript
// ✅ Хорошо - одна ответственность
async function createUser(dto: CreateUserDto): Promise<Result<User, AppError>> {
  const validation = validateUser(dto);
  if (validation.isErr()) return validation;

  const user = await userRepo.create(validation.value);
  if (user.isErr()) return user;

  await sendWelcomeEmail(user.value);
  return user;
}

// ❌ Плохо - слишком много
async function processUserRequest(event: HttpEvent) {
  // 200 строк кода...
}
```

### Early Return

```typescript
// ✅ Хорошо - early return
function processUser(user: User | null): Result<User, AppError> {
  if (!user) {
    return err(createNotFoundError('User not found'));
  }

  if (!user.isActive) {
    return err(createValidationError('User is inactive'));
  }

  return ok(user);
}

// ❌ Плохо - вложенные if
function processUser(user: User | null): Result<User, AppError> {
  if (user) {
    if (user.isActive) {
      return ok(user);
    } else {
      return err(createValidationError('User is inactive'));
    }
  } else {
    return err(createNotFoundError('User not found'));
  }
}
```

## Комментарии

```typescript
// ✅ Хорошо - объясняет ПОЧЕМУ
// Используем Math.floor вместо parseInt для производительности
// в hot path с миллионами вызовов
const index = Math.floor(value);

// ✅ Хорошо - TODO с контекстом
// TODO(#123): Добавить кэширование после оптимизации YDB
const users = await fetchUsers();

// ❌ Плохо - объясняет ЧТО (код и так читаемый)
// Получаем пользователя по ID
const user = await getUser(id);
```

## Git Workflow

### Ветки

```
main              # Продакшен
develop           # Разработка (если есть)
feature/xxx       # Новый функционал
fix/xxx           # Баг фикс
refactor/xxx      # Рефакторинг
```

### Коммиты (Conventional Commits)

```
feat: add user authentication
fix: resolve race condition in cache
refactor: simplify error handling
docs: update API documentation
test: add unit tests for UserService
chore: update dependencies
ci: fix GitLab pipeline
```

### Пример коммита

```
feat: add user registration endpoint

- Add POST /api/v1/users endpoint
- Implement email validation
- Add rate limiting (10 req/min)

Closes #42
```

## Code Review Checklist

### Функциональность
- [ ] Код делает то, что заявлено
- [ ] Edge cases обработаны
- [ ] Ошибки обрабатываются (Result types)

### Качество кода
- [ ] Понятные имена переменных и функций
- [ ] Нет дублирования
- [ ] Функции небольшие и сфокусированные
- [ ] TypeScript strict mode проходит

### Безопасность
- [ ] Нет секретов в коде
- [ ] Входные данные валидируются
- [ ] SQL параметризован

### Тесты
- [ ] Есть тесты для нового кода
- [ ] Тесты проходят
- [ ] Coverage не упал

### Документация
- [ ] Публичные API задокументированы
- [ ] Сложная логика прокомментирована

## PR Template

```markdown
## Описание
Краткое описание изменений

## Тип изменения
- [ ] Новый функционал
- [ ] Баг фикс
- [ ] Рефакторинг
- [ ] Документация

## Как тестировать
1. Шаг 1
2. Шаг 2
3. Ожидаемый результат

## Checklist
- [ ] Код соответствует стандартам
- [ ] Тесты добавлены/обновлены
- [ ] Документация обновлена
- [ ] Self-review проведён

## Related Issues
Closes #XX
```

## Biome Configuration

```json
{
  "linter": {
    "rules": {
      "recommended": true,
      "correctness": {
        "noUnusedImports": "warn",
        "noUnusedVariables": "warn"
      },
      "style": {
        "useConst": "error",
        "noNonNullAssertion": "warn"
      }
    }
  },
  "formatter": {
    "indentStyle": "tab",
    "lineWidth": 100
  },
  "javascript": {
    "formatter": {
      "quoteStyle": "single",
      "semicolons": "always"
    }
  }
}
```

## Best Practices Summary

1. **neverthrow** для ошибок, не try-catch
2. **Biome** для форматирования
3. **TypeScript strict** mode
4. **Conventional commits**
5. **Early return** вместо вложенных if
6. **Маленькие функции** с одной ответственностью
7. **Тесты** для нового кода
8. **Code review** перед мержем
