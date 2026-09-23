# careTech backend architecture

## Project structure

- cmd/server/main.go - entry point, HTTP server startup
- internal/httpapi/handler.go - HTTP routes and JSON handlers
- internal/catalog/product.go - EKT-aligned product contract
- internal/catalog/memory.go - in-memory catalog, search, alternatives
- internal/cart/cart.go - cart store and quantity validation
- internal/chat/service.go - chat session logic, pending confirmation, catalog-first flow
- internal/llm/client.go - NVIDIA API client
- internal/db/postgres.go - PostgreSQL connectivity and schema bootstrap
- .env.example - environment variables template
- nvidia_setup.md - setup instructions

## What the backend does

1. Checks service health via GET /health
2. Searches catalog via GET /api/products?q=...
3. Handles user chat via POST /api/chat
4. Supports pending product confirmation before adding to cart
5. Stores cart state per session_id
6. Confirms cart updates through POST /api/cart/confirm
7. Optionally uses NVIDIA LLM for conversational responses when configured
8. Supports PostgreSQL connection as a DB-ready layer, with in-memory fallback

## Main business flow

- User sends a product query
- Backend normalizes query and searches exact identifiers first
- If a product is found, it checks stock and price
- If product is out of stock, it suggests alternatives
- The product is not inserted into cart until explicit confirmation
- The cart can then be queried and confirmed

## Environment

Required variables:
- NVIDIA_API_KEY
- NVIDIA_API_BASE_URL
- NVIDIA_MODEL
- DATABASE_URL

## Verification status

- Go tests pass
- Project compiles successfully
- The backend remains in MVP mode, with in-memory storage still used as the primary fallback
