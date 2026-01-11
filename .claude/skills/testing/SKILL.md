---
name: testing
description: Тестирование в @mrg0773 с Vitest. Unit тесты, моки, coverage. Используй для написания тестов, настройки Vitest, мокирования.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# Testing с Vitest

## Конфигурация

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    include: ['**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts}'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: ['node_modules', 'dist', '**/*.d.ts']
    }
  }
});
```

## Команды

```bash
# Запуск тестов
npm test

# Watch mode
npm run test:watch

# С coverage
npm run test:coverage

# Конкретный файл
npx vitest run src/services/user.test.ts
```

## Базовый тест

```typescript
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { UserService } from './user.service';

describe('UserService', () => {
  let service: UserService;

  beforeEach(() => {
    service = new UserService();
  });

  afterEach(() => {
    // Cleanup
  });

  it('should create user', async () => {
    const result = await service.create({ name: 'John' });

    expect(result.isOk()).toBe(true);
    expect(result.value).toMatchObject({
      name: 'John'
    });
  });

  it('should return error for invalid input', async () => {
    const result = await service.create({ name: '' });

    expect(result.isErr()).toBe(true);
    expect(result.error.code).toBe('VALIDATION_ERROR');
  });
});
```

## Тестирование Result типов

```typescript
import { ok, err, Result } from '@mrg0773/logger-error';

describe('Result handling', () => {
  it('should handle success', () => {
    const result = ok({ id: 1 });

    expect(result.isOk()).toBe(true);
    if (result.isOk()) {
      expect(result.value.id).toBe(1);
    }
  });

  it('should handle error', () => {
    const result = err(createValidationError('Invalid'));

    expect(result.isErr()).toBe(true);
    if (result.isErr()) {
      expect(result.error.code).toBe('VALIDATION_ERROR');
    }
  });
});
```

## Мокирование

### vi.fn() - функции

```typescript
import { vi } from 'vitest';

const mockFn = vi.fn();
mockFn.mockReturnValue('mocked');

expect(mockFn()).toBe('mocked');
expect(mockFn).toHaveBeenCalled();
```

### vi.mock() - модули

```typescript
import { vi } from 'vitest';

// Мокаем модуль
vi.mock('@mrg0773/ydb', () => ({
  YDBGlobalConnection: {
    initialize: vi.fn().mockResolvedValue(ok(undefined)),
    getYDB: vi.fn().mockReturnValue({
      query: vi.fn().mockResolvedValue(ok([]))
    })
  }
}));

// Тест
it('should use mocked YDB', async () => {
  await YDBGlobalConnection.initialize('test');
  const ydb = YDBGlobalConnection.getYDB();

  const result = await ydb.query`SELECT 1`;
  expect(result.isOk()).toBe(true);
});
```

### vi.spyOn() - шпионы

```typescript
import { vi } from 'vitest';

const spy = vi.spyOn(service, 'getUser');
spy.mockResolvedValue(ok({ id: 1, name: 'John' }));

const result = await service.getUser('1');

expect(spy).toHaveBeenCalledWith('1');
expect(result.isOk()).toBe(true);
```

## Асинхронные тесты

```typescript
describe('Async operations', () => {
  it('should resolve', async () => {
    const result = await asyncOperation();
    expect(result).toBe('done');
  });

  it('should reject', async () => {
    await expect(failingOperation()).rejects.toThrow('Error');
  });

  it('should timeout', async () => {
    await expect(
      asyncOperation()
    ).resolves.toBe('done');
  }, 5000); // Custom timeout
});
```

## Snapshot тесты

```typescript
it('should match snapshot', () => {
  const output = generateReport();
  expect(output).toMatchSnapshot();
});

it('should match inline snapshot', () => {
  const output = formatError(error);
  expect(output).toMatchInlineSnapshot(`
    "Error: Something went wrong
    Code: VALIDATION_ERROR"
  `);
});
```

## Test Fixtures

```typescript
// fixtures/users.ts
export const testUser = {
  id: '1',
  name: 'Test User',
  email: 'test@example.com'
};

export const testUsers = [
  testUser,
  { id: '2', name: 'Another User', email: 'another@example.com' }
];

// test file
import { testUser, testUsers } from './fixtures/users';

it('should process user', () => {
  const result = processUser(testUser);
  expect(result.name).toBe('Test User');
});
```

## Тестирование репозиториев

```typescript
describe('UserRepository', () => {
  let repo: UserRepository;
  let mockYdb: MockYDB;

  beforeEach(() => {
    mockYdb = createMockYDB();
    repo = new UserRepository(mockYdb);
  });

  it('should find user by id', async () => {
    mockYdb.query.mockResolvedValue(ok([testUser]));

    const result = await repo.findById('1');

    expect(result.isOk()).toBe(true);
    expect(result.value).toEqual(testUser);
    expect(mockYdb.query).toHaveBeenCalled();
  });

  it('should return null for non-existent user', async () => {
    mockYdb.query.mockResolvedValue(ok([]));

    const result = await repo.findById('999');

    expect(result.isOk()).toBe(true);
    expect(result.value).toBeNull();
  });
});
```

## Тестирование сервисов

```typescript
describe('OrderService', () => {
  let service: OrderService;
  let mockUserRepo: MockUserRepository;
  let mockOrderRepo: MockOrderRepository;

  beforeEach(() => {
    mockUserRepo = createMockUserRepo();
    mockOrderRepo = createMockOrderRepo();
    service = new OrderService(mockUserRepo, mockOrderRepo);
  });

  it('should place order', async () => {
    mockUserRepo.findById.mockResolvedValue(ok({
      id: '1',
      balance: 1000
    }));
    mockOrderRepo.create.mockResolvedValue(ok({
      id: 'order-1',
      total: 100
    }));

    const result = await service.placeOrder({
      userId: '1',
      items: [{ productId: 'p1', quantity: 1 }],
      total: 100
    });

    expect(result.isOk()).toBe(true);
    expect(mockOrderRepo.create).toHaveBeenCalled();
  });

  it('should fail for insufficient balance', async () => {
    mockUserRepo.findById.mockResolvedValue(ok({
      id: '1',
      balance: 50
    }));

    const result = await service.placeOrder({
      userId: '1',
      items: [],
      total: 100
    });

    expect(result.isErr()).toBe(true);
    expect(result.error.code).toBe('VALIDATION_ERROR');
  });
});
```

## Coverage Requirements

```json
// package.json
{
  "scripts": {
    "test:coverage": "vitest run --coverage"
  }
}
```

Минимальные требования @mrg0773:
- **Statements**: 80%
- **Branches**: 75%
- **Functions**: 80%
- **Lines**: 80%

## Best Practices

1. **Один assert на тест** (где возможно)
2. **Describe блоки** для группировки
3. **beforeEach/afterEach** для setup/cleanup
4. **Моки для внешних зависимостей** (YDB, Redis)
5. **Тестируй Result типы** правильно
6. **Fixtures** для тестовых данных
7. **Coverage** как метрика, не цель
