# Order Food API Server

Go implementation of the food ordering API based on the [OpenAPI 3.1 specification](https://orderfoodonline.deno.dev/public/openapi.yaml).

> **Challenge origin**: [oolio-group/kart-challenge](https://github.com/oolio-group/kart-challenge/blob/advanced-challenge/backend-challenge/README.md)

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/product` | No | List all products |
| `GET` | `/product/{productId}` | No | Get a single product by ID |
| `POST` | `/order` | **Yes** | Place a new order |
| `GET` | `/healthz` | No | Health check |

## Authentication

Order endpoint requires the `Api_key` header set to `apitest`:

```
Api_key: apitest
```

## Product Catalog (8 items)

| ID | Name | Price | Category |
|----|------|-------|----------|
| 10 | Chicken Waffle | $13.30 | Waffle |
| 11 | Belgian Waffle | $9.50 | Waffle |
| 12 | Strawberry Waffle | $11.00 | Waffle |
| 20 | Maple Syrup | $3.50 | Topping |
| 21 | Whipped Cream | $2.00 | Topping |
| 22 | Chocolate Sauce | $2.50 | Topping |
| 30 | Fresh Orange Juice | $4.50 | Drink |
| 31 | Iced Coffee | $5.00 | Drink |

## Coupons

Valid promo codes (appear in ≥2 of 3 offline coupon files):

| Code | Discount |
|------|----------|
| `HAPPYHRS` | 18% off order total |
| `FIFTYOFF` | 18% off order total |
| `SIXTYOFF` | 18% off order total |
| `BIRTHDAY` | 18% off order total |
| `GNULINUX` | 18% off order total |
| `OVER9000` | 18% off order total |
| `FREEZAAA` | 18% off order total |
| `BUYGETON` | Lowest-priced item free |

Invalid codes (e.g. `SUPER100`) return a 422 error.

## Request / Response Examples

### Place an order (no coupon)

```bash
curl -X POST http://localhost:8080/order \
  -H "Content-Type: application/json" \
  -H "Api_key: apitest" \
  -d '{"items":[{"productId":"10","quantity":3}]}'
```

```json
{
  "id": "adcb9e67-6aa8-4ba6-a028-bf5afbc66422",
  "total": 39.9,
  "discounts": 0,
  "items": [{"productId": "10", "quantity": 3}],
  "products": [{"id": "10", "name": "Chicken Waffle", "price": 13.3, "category": "Waffle", "image": {...}}]
}
```

### Place an order (18% coupon)

```bash
curl -X POST http://localhost:8080/order \
  -H "Content-Type: application/json" \
  -H "Api_key: apitest" \
  -d '{"couponCode":"HAPPYHRS","items":[{"productId":"10","quantity":1}]}'
```

```json
{
  "id": "a0422ca0-706a-4024-bb1e-f5b30c114fa2",
  "total": 10.91,
  "discounts": 2.39,
  "items": [{"productId": "10", "quantity": 1}],
  "products": [...]
}
```

### Place an order (BUYGETON — cheapest item free)

```bash
curl -X POST http://localhost:8080/order \
  -H "Content-Type: application/json" \
  -H "Api_key: apitest" \
  -d '{"couponCode":"BUYGETON","items":[{"productId":"10","quantity":1},{"productId":"30","quantity":1}]}'
```

```json
{
  "id": "90682026-f107-4fea-939e-57edc11b1207",
  "total": 13.3,
  "discounts": 4.5,
  "items": [{"productId": "10", "quantity": 1}, {"productId": "30", "quantity": 1}],
  "products": [...]
}
```

## Error Responses

| Status | Scenario |
|--------|----------|
| 401 | Missing or invalid `Api_key` header |
| 404 | Product not found |
| 422 | Invalid input (empty items, zero quantity, unknown product, invalid coupon) |

## Running

```bash
go build -o orderfood .
./orderfood            # starts on :8080
PORT=3000 ./orderfood  # starts on :3000
```

Graceful shutdown on `SIGINT` / `SIGTERM`.

## Tests

```bash
go test ./...
# 21 tests, all passing
```

## Architecture

```
├── go.mod
├── main.go                          # Server bootstrap
└── internal/
    ├── coupon/
    │   └── coupon.go                # Valid codes set + discount logic
    ├── handler/
    │   ├── product.go               # GET /product endpoints
    │   └── order.go                 # POST /order + UUID v4
    ├── middleware/
    │   └── auth.go                  # API key validation
    └── model/
        ├── product.go               # Product structs + catalog
        └── order.go                 # Order structs + validation
    └── test/
        └── handler_test.go          # 21 unit + integration tests
```

## Zero Dependencies

Built entirely with Go standard library — `net/http`, `encoding/json`, `crypto/rand`, `compress/gzip`, etc. No external packages.