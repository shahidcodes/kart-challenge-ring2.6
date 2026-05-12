package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"math"
	"net/http"
	"strings"

	"orderfood/internal/coupon"
	"orderfood/internal/model"
)

// PlaceOrder handles POST /order.
func PlaceOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "Method not allowed",
		})
		return
	}

	// Decode request body with strict validation
	var req model.OrderReq
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate the request
	if msg := req.Validate(); msg != "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{
			"error": msg,
		})
		return
	}

	// Build order lines, look up each product
	var lines []model.OrderLine
	var products []model.Product
	var subtotal float64

	for _, item := range req.Items {
		p := model.Lookup(item.ProductID)
		if p == nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{
				"error": "Product not found: " + item.ProductID,
			})
			return
		}
		lines = append(lines, model.OrderLine{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
		products = append(products, *p)
		subtotal += p.Price * float64(item.Quantity)
	}

	// Apply coupon if provided
	var discount float64
	couponCode := strings.TrimSpace(req.CouponCode)
	if couponCode != "" {
		if !coupon.Valid(couponCode) {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{
				"error": "Invalid coupon code",
			})
			return
		}
		info := coupon.Info(couponCode)
		if info != nil {
			switch info.Type {
			case coupon.DiscountPercent:
				discount = round2(subtotal * info.Percent / 100)
			case coupon.DiscountCheapestFree:
				discount = cheapestItemPrice(req.Items)
			}
		}
	}

	total := round2(subtotal - discount)
	if total < 0 {
		total = 0
	}

	// Generate UUID v4 for order ID
	orderID := uuid4()

	order := model.Order{
		ID:        orderID,
		Total:     total,
		Discounts: discount,
		Items:     lines,
		Products:  products,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(order)
}

// cheapestItemPrice returns the price of the cheapest product in the items list.
func cheapestItemPrice(items []model.OrderItem) float64 {
	if len(items) == 0 {
		return 0
	}
	var minPrice float64 = math.MaxFloat64
	for _, item := range items {
		p := model.Lookup(item.ProductID)
		if p != nil && p.Price < minPrice {
			minPrice = p.Price
		}
	}
	if minPrice == math.MaxFloat64 {
		return 0
	}
	return minPrice
}

// round2 rounds a float64 to 2 decimal places.
func round2(f float64) float64 {
	return math.Round(f*100) / 100
}

// uuid4 generates a random UUID v4 string.
func uuid4() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(err) // crypto/rand should never fail
	}
	// Set version (4) and variant (RFC 4122)
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80

	return hex.EncodeToString(buf[:4]) + "-" +
		hex.EncodeToString(buf[4:6]) + "-" +
		hex.EncodeToString(buf[6:8]) + "-" +
		hex.EncodeToString(buf[8:10]) + "-" +
		hex.EncodeToString(buf[10:])
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(v)
}