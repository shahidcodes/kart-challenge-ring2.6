package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"orderfood/internal/coupon"
	"orderfood/internal/handler"
	"orderfood/internal/middleware"
	"orderfood/internal/model"
)

// ==================== Product Handler Tests ====================

func TestListProducts(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/product", nil)
	w := httptest.NewRecorder()

	handler.ListProducts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var products []model.Product
	if err := json.NewDecoder(w.Body).Decode(&products); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("expected at least one product")
	}

	// Verify Chicken Waffle is present
	found := false
	for _, p := range products {
		if p.ID == "10" && p.Name == "Chicken Waffle" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Chicken Waffle (id=10) in catalog")
	}
}

func TestGetProduct(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"valid product", "/product/10", http.StatusOK},
		{"another product", "/product/11", http.StatusOK},
		{"nonexistent product", "/product/999", http.StatusNotFound},
		{"empty id", "/product/", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			handler.GetProduct(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if tt.wantStatus == http.StatusOK {
				var p model.Product
				if err := json.NewDecoder(w.Body).Decode(&p); err != nil {
					t.Fatalf("failed to decode: %v", err)
				}
				if p.ID == "" {
					t.Error("expected non-empty product")
				}
			}
		})
	}
}

// ==================== Order Handler Tests ====================

func makeOrderReq(couponCode string, items []model.OrderItem) *http.Request {
	body := model.OrderReq{CouponCode: couponCode, Items: items}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/order", strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(middleware.APIKeyHeader, middleware.ExpectedAPIKey)
	return req
}

func TestPlaceOrderUnauthenticated(t *testing.T) {
	mux := buildMux()
	req := httptest.NewRequest(http.MethodPost, "/order", strings.NewReader(`{"items":[{"productId":"10","quantity":1}]}`))
	req.Header.Set("Content-Type", "application/json")
	// No API key header
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestPlaceOrderNoItems(t *testing.T) {
	req := makeOrderReq("", []model.OrderItem{})
	w := httptest.NewRecorder()
	handler.PlaceOrder(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

func TestPlaceOrderInvalidProduct(t *testing.T) {
	req := makeOrderReq("", []model.OrderItem{{ProductID: "9999", Quantity: 1}})
	w := httptest.NewRecorder()
	handler.PlaceOrder(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

func TestPlaceOrderInvalidQuantity(t *testing.T) {
	req := makeOrderReq("", []model.OrderItem{{ProductID: "10", Quantity: 0}})
	w := httptest.NewRecorder()
	handler.PlaceOrder(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

func TestPlaceOrderNoCoupon(t *testing.T) {
	req := makeOrderReq("", []model.OrderItem{{ProductID: "10", Quantity: 2}})
	w := httptest.NewRecorder()
	handler.PlaceOrder(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var order model.Order
	if err := json.NewDecoder(w.Body).Decode(&order); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	// 2 × $13.30 = $26.60, no discount
	if order.Total != 26.6 {
		t.Errorf("expected total 26.6, got %.2f", order.Total)
	}
	if order.Discounts != 0 {
		t.Errorf("expected 0 discounts, got %.2f", order.Discounts)
	}
	if len(order.Items) != 1 {
		t.Fatalf("expected 1 order item, got %d", len(order.Items))
	}
	if order.Items[0].Quantity != 2 {
		t.Errorf("expected quantity 2, got %d", order.Items[0].Quantity)
	}
	if len(order.ID) == 0 {
		t.Error("expected non-empty order ID")
	}
}

func TestPlaceOrderWithPercentCoupon(t *testing.T) {
	// HAPPYHRS → 18% off
	req := makeOrderReq("HAPPYHRS", []model.OrderItem{{ProductID: "10", Quantity: 1}})
	w := httptest.NewRecorder()
	handler.PlaceOrder(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var order model.Order
	if err := json.NewDecoder(w.Body).Decode(&order); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	// $13.30 × 0.18 = $2.394 → round to $2.39
	// total = 13.30 - 2.39 = 10.91
	if order.Discounts != 2.39 {
		t.Errorf("expected discount 2.39, got %.2f", order.Discounts)
	}
	if order.Total != 10.91 {
		t.Errorf("expected total 10.91, got %.2f", order.Total)
	}
}

func TestPlaceOrderWithInvalidCoupon(t *testing.T) {
	req := makeOrderReq("INVALIDCODE", []model.OrderItem{{ProductID: "10", Quantity: 1}})
	w := httptest.NewRecorder()
	handler.PlaceOrder(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

func TestPlaceOrderBuyOneGetOne(t *testing.T) {
	// BUYGETON → lowest-priced item free
	req := makeOrderReq("BUYGETON", []model.OrderItem{
		{ProductID: "10", Quantity: 1}, // $13.30
		{ProductID: "30", Quantity: 1}, // $4.50 (cheapest)
	})
	w := httptest.NewRecorder()
	handler.PlaceOrder(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var order model.Order
	if err := json.NewDecoder(w.Body).Decode(&order); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	// subtotal = 13.30 + 4.50 = 17.80, discount = 4.50
	// total = 13.30
	if order.Discounts != 4.50 {
		t.Errorf("expected discount 4.50, got %.2f", order.Discounts)
	}
	if order.Total != 13.30 {
		t.Errorf("expected total 13.30, got %.2f", order.Total)
	}
}

func TestPlaceOrderMultipleSameProduct(t *testing.T) {
	req := makeOrderReq("", []model.OrderItem{{ProductID: "10", Quantity: 3}})
	w := httptest.NewRecorder()
	handler.PlaceOrder(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var order model.Order
	if err := json.NewDecoder(w.Body).Decode(&order); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	// 3 × $13.30 = $39.90
	if order.Total != 39.90 {
		t.Errorf("expected total 39.90, got %.2f", order.Total)
	}
	if order.Items[0].Quantity != 3 {
		t.Errorf("expected quantity 3, got %d", order.Items[0].Quantity)
	}
}

// ==================== Coupon Tests ====================

func TestCouponValid(t *testing.T) {
	for _, code := range []string{"HAPPYHRS", "BUYGETON", "FIFTYOFF", "SIXTYOFF"} {
		if !coupon.Valid(code) {
			t.Errorf("expected %q to be valid", code)
		}
	}
}

func TestCouponInvalid(t *testing.T) {
	for _, code := range []string{"INVALID", "SUPER100", "TOOSHORT", "TOOLOOOOOONNNNGGG"} {
		if coupon.Valid(code) {
			t.Errorf("expected %q to be invalid", code)
		}
	}
}

func TestCouponCaseInsensitive(t *testing.T) {
	if !coupon.Valid("happyhrs") {
		t.Error("expected lowercase happyhrs to be valid")
	}
	if !coupon.Valid("HappyHrs") {
		t.Error("expected mixed-case HappyHrs to be valid")
	}
}

func TestCouponInfo(t *testing.T) {
	info := coupon.Info("HAPPYHRS")
	if info == nil {
		t.Fatal("expected non-nil info for HAPPYHRS")
	}
	if info.Type != coupon.DiscountPercent {
		t.Errorf("expected DiscountPercent, got %v", info.Type)
	}
	if info.Percent != 18 {
		t.Errorf("expected 18%%, got %.0f%%", info.Percent)
	}
}

func TestCouponInfoBuyGetOn(t *testing.T) {
	info := coupon.Info("BUYGETON")
	if info == nil {
		t.Fatal("expected non-nil info for BUYGETON")
	}
	if info.Type != coupon.DiscountCheapestFree {
		t.Errorf("expected DiscountCheapestFree, got %v", info.Type)
	}
}

func TestCouponInfoInvalid(t *testing.T) {
	if coupon.Info("INVALID") != nil {
		t.Error("expected nil info for invalid coupon")
	}
}

// ==================== Router Integration Tests ====================

func TestRouterProductEndpoints(t *testing.T) {
	mux := buildMux()

	t.Run("GET /product returns list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/product", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("GET /product/10 returns product", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/product/10", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var p model.Product
		json.NewDecoder(w.Body).Decode(&p)
		if p.ID != "10" {
			t.Errorf("expected product id 10, got %s", p.ID)
		}
	})

	t.Run("GET /product/999 returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/product/999", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}

func TestRouterOrderEndpoint(t *testing.T) {
	mux := buildMux()

	t.Run("POST /order without API key returns 401", func(t *testing.T) {
		body := model.OrderReq{Items: []model.OrderItem{{ProductID: "10", Quantity: 1}}}
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/order", strings.NewReader(string(b)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("POST /order with valid API key succeeds", func(t *testing.T) {
		body := model.OrderReq{Items: []model.OrderItem{{ProductID: "10", Quantity: 1}}}
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/order", strings.NewReader(string(b)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(middleware.APIKeyHeader, middleware.ExpectedAPIKey)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}

func buildMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /product", handler.ListProducts)
	mux.HandleFunc("GET /product/{productId}", handler.GetProduct)
	mux.HandleFunc("POST /order", middleware.Auth(handler.PlaceOrder))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})
	return mux
}