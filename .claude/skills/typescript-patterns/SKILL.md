---
name: typescript-patterns
description: TypeScript паттерны в @mrg0773. Strict mode, типы, дженерики, type guards. Используй для типизации, создания типов, работы с TypeScript.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# TypeScript Patterns

## Конфигурация

Стандартный tsconfig.json для @mrg0773:

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "exactOptionalPropertyTypes": true,
    "noUncheckedIndexedAccess": true,
    "noPropertyAccessFromIndexSignature": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "outDir": "dist",
    "rootDir": "src"
  }
}
```

## Strict Mode Checks

### strictNullChecks

```typescript
// ❌ Без strict - ошибка в runtime
function getLength(str: string) {
  return str.length; // str может быть null!
}

// ✅ Со strict - ошибка компиляции
function getLength(str: string | null) {
  if (str === null) return 0;
  return str.length;
}
```

### noUncheckedIndexedAccess

```typescript
const arr = [1, 2, 3];

// ❌ Без флага
const first = arr[0]; // type: number

// ✅ С флагом
const first = arr[0]; // type: number | undefined
if (first !== undefined) {
  console.log(first); // type: number
}
```

### exactOptionalPropertyTypes

```typescript
interface User {
  name: string;
  email?: string;
}

// ❌ Без флага - можно присвоить undefined
const user: User = { name: 'John', email: undefined };

// ✅ С флагом - нужно опустить свойство
const user: User = { name: 'John' };
// или
const user: User = { name: 'John', email: 'john@example.com' };
```

## Type Guards

### typeof

```typescript
function process(value: string | number) {
  if (typeof value === 'string') {
    return value.toUpperCase(); // value: string
  }
  return value * 2; // value: number
}
```

### instanceof

```typescript
function handleError(error: Error | AppError) {
  if (error instanceof AppError) {
    console.log(error.code); // AppError specific
  } else {
    console.log(error.message); // Generic Error
  }
}
```

### Custom Type Guards

```typescript
interface User {
  type: 'user';
  name: string;
}

interface Admin {
  type: 'admin';
  name: string;
  permissions: string[];
}

// Type predicate
function isAdmin(entity: User | Admin): entity is Admin {
  return entity.type === 'admin';
}

function getPermissions(entity: User | Admin) {
  if (isAdmin(entity)) {
    return entity.permissions; // entity: Admin
  }
  return []; // entity: User
}
```

### Discriminated Unions

```typescript
type Result<T, E> =
  | { success: true; value: T }
  | { success: false; error: E };

function handle<T, E>(result: Result<T, E>) {
  if (result.success) {
    console.log(result.value); // value доступен
  } else {
    console.log(result.error); // error доступен
  }
}
```

## Branded Types

Предотвращение смешивания типов:

```typescript
// Бренды
type UserId = string & { readonly __brand: 'UserId' };
type OrderId = string & { readonly __brand: 'OrderId' };

// Фабрики
function createUserId(id: string): UserId {
  return id as UserId;
}

function createOrderId(id: string): OrderId {
  return id as OrderId;
}

// Использование
function getUser(id: UserId) { /* ... */ }
function getOrder(id: OrderId) { /* ... */ }

const userId = createUserId('123');
const orderId = createOrderId('456');

getUser(userId);  // ✅ OK
getUser(orderId); // ❌ Type error!
```

## Utility Types

### Partial / Required

```typescript
interface User {
  id: string;
  name: string;
  email: string;
}

// Все поля опциональны
type UpdateUser = Partial<User>;

// Все поля обязательны
type CompleteUser = Required<User>;
```

### Pick / Omit

```typescript
// Только выбранные поля
type UserPreview = Pick<User, 'id' | 'name'>;

// Все кроме указанных
type CreateUser = Omit<User, 'id'>;
```

### Record

```typescript
// Объект с ключами определённого типа
type ErrorMessages = Record<ErrorCode, string>;

const messages: ErrorMessages = {
  VALIDATION_ERROR: 'Invalid input',
  DATABASE_ERROR: 'Database unavailable',
  // ...
};
```

### Extract / Exclude

```typescript
type AllCodes = 'A' | 'B' | 'C' | 'D';

// Только A и B
type Selected = Extract<AllCodes, 'A' | 'B'>; // 'A' | 'B'

// Все кроме A и B
type Remaining = Exclude<AllCodes, 'A' | 'B'>; // 'C' | 'D'
```

## Generics

### Basic

```typescript
function identity<T>(value: T): T {
  return value;
}

const str = identity('hello'); // type: string
const num = identity(42);      // type: number
```

### Constraints

```typescript
interface HasId {
  id: string;
}

function getById<T extends HasId>(items: T[], id: string): T | undefined {
  return items.find(item => item.id === id);
}
```

### Multiple Type Parameters

```typescript
function map<T, U>(items: T[], fn: (item: T) => U): U[] {
  return items.map(fn);
}

const numbers = [1, 2, 3];
const strings = map(numbers, n => n.toString()); // string[]
```

### Default Type Parameters

```typescript
interface ApiResponse<T = unknown> {
  data: T;
  status: number;
}

const response: ApiResponse = { data: 'test', status: 200 };
const typedResponse: ApiResponse<User> = { data: user, status: 200 };
```

## Conditional Types

```typescript
type IsString<T> = T extends string ? true : false;

type A = IsString<string>;  // true
type B = IsString<number>;  // false

// Практический пример
type ArrayElement<T> = T extends (infer U)[] ? U : never;

type Elem = ArrayElement<string[]>; // string
```

## Mapped Types

```typescript
// Сделать все поля readonly
type Readonly<T> = {
  readonly [K in keyof T]: T[K];
};

// Сделать все поля nullable
type Nullable<T> = {
  [K in keyof T]: T[K] | null;
};

// Добавить префикс к ключам
type Prefixed<T, P extends string> = {
  [K in keyof T as `${P}${string & K}`]: T[K];
};
```

## Template Literal Types

```typescript
type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE';
type Endpoint = '/users' | '/orders';

// Комбинации
type Route = `${HttpMethod} ${Endpoint}`;
// 'GET /users' | 'GET /orders' | 'POST /users' | ...
```

## Assertion Functions

```typescript
function assertIsString(value: unknown): asserts value is string {
  if (typeof value !== 'string') {
    throw new Error('Value is not a string');
  }
}

function process(value: unknown) {
  assertIsString(value);
  // После assertion value: string
  console.log(value.toUpperCase());
}
```

## Best Practices @mrg0773

1. **Strict mode всегда включён**
2. **Branded types** для ID сущностей
3. **Type guards** вместо type assertions
4. **Discriminated unions** для Result типов
5. **Utility types** вместо дублирования
6. **Generics** для переиспользуемого кода
7. **Все типы экспортируются** из @mrg0773/types
8. **95%+ type coverage** requirement
