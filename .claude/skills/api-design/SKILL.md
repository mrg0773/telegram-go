---
name: api-design
description: Проектирование REST API в @mrg0773. Endpoints, contracts, версионирование, документация. Используй для создания API, контрактов, OpenAPI.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write
---

# API Design

## Структура Endpoint

```
METHOD /api/v1/resource/:id/subresource
```

| Метод | Назначение | Пример |
|-------|-----------|--------|
| GET | Получить | GET /api/v1/users/123 |
| POST | Создать | POST /api/v1/users |
| PUT | Заменить | PUT /api/v1/users/123 |
| PATCH | Обновить | PATCH /api/v1/users/123 |
| DELETE | Удалить | DELETE /api/v1/users/123 |

## Response Format

### Успешный ответ

```typescript
// Один объект
{
  "data": {
    "id": "123",
    "name": "John",
    "email": "john@example.com"
  }
}

// Коллекция
{
  "data": [
    { "id": "1", "name": "John" },
    { "id": "2", "name": "Jane" }
  ],
  "meta": {
    "total": 100,
    "page": 1,
    "perPage": 10
  }
}
```

### Ошибка

```typescript
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ]
  }
}
```

## HTTP Status Codes

| Code | Значение | Когда использовать |
|------|----------|-------------------|
| 200 | OK | Успешный GET, PUT, PATCH |
| 201 | Created | Успешный POST |
| 204 | No Content | Успешный DELETE |
| 400 | Bad Request | Ошибка валидации |
| 401 | Unauthorized | Не авторизован |
| 403 | Forbidden | Нет прав доступа |
| 404 | Not Found | Ресурс не найден |
| 409 | Conflict | Конфликт (duplicate) |
| 429 | Too Many Requests | Rate limit |
| 500 | Internal Error | Ошибка сервера |

## Handler Template

```typescript
import { ok, err, Result } from '@mrg0773/logger-error';

interface HttpEvent {
  httpMethod: string;
  pathParameters: Record<string, string>;
  queryStringParameters: Record<string, string>;
  body: string;
  headers: Record<string, string>;
}

interface HttpResponse {
  statusCode: number;
  headers: Record<string, string>;
  body: string;
}

export async function handler(event: HttpEvent): Promise<HttpResponse> {
  try {
    // 1. Parse input
    const body = event.body ? JSON.parse(event.body) : {};
    const params = event.pathParameters || {};
    const query = event.queryStringParameters || {};

    // 2. Validate
    const validation = validateInput(body);
    if (validation.isErr()) {
      return errorResponse(400, validation.error);
    }

    // 3. Process
    const result = await processRequest(validation.value);
    if (result.isErr()) {
      return errorResponse(getStatusCode(result.error), result.error);
    }

    // 4. Response
    return successResponse(200, result.value);

  } catch (error) {
    return errorResponse(500, handleUnknownError(error));
  }
}

function successResponse(statusCode: number, data: unknown): HttpResponse {
  return {
    statusCode,
    headers: {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*'
    },
    body: JSON.stringify({ data })
  };
}

function errorResponse(statusCode: number, error: AppError): HttpResponse {
  return {
    statusCode,
    headers: {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*'
    },
    body: JSON.stringify({
      error: {
        code: error.code,
        message: error.message
      }
    })
  };
}

function getStatusCode(error: AppError): number {
  const statusMap: Record<ErrorCode, number> = {
    VALIDATION_ERROR: 400,
    AUTH_ERROR: 401,
    NOT_FOUND_ERROR: 404,
    DATABASE_ERROR: 500,
    API_ERROR: 502,
    TIMEOUT_ERROR: 504
  };
  return statusMap[error.code] || 500;
}
```

## OpenAPI / Swagger

```yaml
openapi: 3.0.0
info:
  title: TGFerma API
  version: 1.0.0

servers:
  - url: https://api.example.com/v1

paths:
  /users:
    get:
      summary: List users
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            default: 1
        - name: limit
          in: query
          schema:
            type: integer
            default: 10
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/User'
                  meta:
                    $ref: '#/components/schemas/Pagination'

    post:
      summary: Create user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUser'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'

  /users/{id}:
    get:
      summary: Get user by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: Not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        email:
          type: string
      required: [id, name, email]

    CreateUser:
      type: object
      properties:
        name:
          type: string
        email:
          type: string
      required: [name, email]

    Pagination:
      type: object
      properties:
        total:
          type: integer
        page:
          type: integer
        perPage:
          type: integer

    Error:
      type: object
      properties:
        error:
          type: object
          properties:
            code:
              type: string
            message:
              type: string
```

## Версионирование

```
/api/v1/users  → текущая версия
/api/v2/users  → новая версия (breaking changes)
```

### В заголовках

```
Accept: application/vnd.myapi.v2+json
```

## Пагинация

### Offset-based

```typescript
GET /api/v1/users?page=2&limit=10

{
  "data": [...],
  "meta": {
    "total": 100,
    "page": 2,
    "perPage": 10,
    "totalPages": 10
  }
}
```

### Cursor-based (для больших данных)

```typescript
GET /api/v1/users?cursor=abc123&limit=10

{
  "data": [...],
  "meta": {
    "nextCursor": "def456",
    "hasMore": true
  }
}
```

## Фильтрация и сортировка

```typescript
// Фильтрация
GET /api/v1/users?status=active&role=admin

// Сортировка
GET /api/v1/users?sort=name&order=asc

// Поиск
GET /api/v1/users?search=john
```

## CORS

```typescript
const corsHeaders = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Methods': 'GET, POST, PUT, PATCH, DELETE, OPTIONS',
  'Access-Control-Allow-Headers': 'Content-Type, Authorization',
  'Access-Control-Max-Age': '86400'
};

// Preflight
if (event.httpMethod === 'OPTIONS') {
  return {
    statusCode: 204,
    headers: corsHeaders,
    body: ''
  };
}
```

## Best Practices

1. **Consistent naming** - plural nouns (users, orders)
2. **Versioning** - /api/v1/...
3. **Proper status codes** - не всегда 200
4. **Error format** - единый формат ошибок
5. **Pagination** - для коллекций
6. **Rate limiting** - защита от abuse
7. **Documentation** - OpenAPI spec
8. **Idempotency** - для POST/PUT
