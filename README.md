# EKT Navigator React

Frontend MVP для ИИ-консультанта EKT. Чат является главным сценарием: запрос отправляется в Go через `/api/chat`, а карточки, аналоги, наличие и корзина используют тот же серверный контракт.

## Запуск

```bash
npm install
npm run dev
```

Для API-режима запустите Go backend во втором терминале на `http://localhost:8080`. Vite проксирует `/api` на этот адрес.

```bash
go run ./...
```

Откройте `http://localhost:5173` и начните с вопроса `Есть ли 515291?`.

## API-контракт

- `GET /api/search?q=...` -> `{ "items": [ProductSummary] }`
- `GET /api/products/{id}` -> `Product`
- `GET /api/products/{id}/alternatives` -> `{ "source_product_id": "...", "items": [{ "product": Product, "reason": "...", "differences": [] }] }`
- `POST /api/chat` -> `{ "reply": "...", "products": [], "pending_action": null | { "product_id", "quantity", "confirmation_token" } }`
- `POST /api/cart/confirm` с `{ "confirmation_token": "..." }` -> `{ "status", "mode", "cart_url?" }`; `409` означает, что цена или остаток изменились.

В `Product` значение `quantity: null` означает неизвестное наличие, а `0` означает подтверждённое отсутствие. Ошибки должны возвращаться как JSON `{ "error": "понятное сообщение" }`.

## Демо-режим

Для показа без Go можно переключить режим в интерфейсе на `Демо-режим`. Демо-данные явно помечены и не называются актуальной корзиной EKT. В API-режиме фронт не обращается к EKT API или LLM напрямую и не хранит ключи в браузере.

## Проверка

```bash
npm run lint
npm run build
```
