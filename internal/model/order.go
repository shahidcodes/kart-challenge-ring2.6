package model

// OrderItem represents a line item in an order request.
type OrderItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// Validate returns an error message if the item is invalid.
func (i OrderItem) Validate() string {
	if i.ProductID == "" {
		return "productId is required"
	}
	if i.Quantity <= 0 {
		return "quantity must be a positive integer"
	}
	return ""
}

// OrderReq is the request body for POST /order.
type OrderReq struct {
	CouponCode string     `json:"couponCode,omitempty"`
	Items      []OrderItem `json:"items"`
}

// Validate returns an error message if the request is invalid.
func (r OrderReq) Validate() string {
	if len(r.Items) == 0 {
		return "items must not be empty"
	}
	for _, item := range r.Items {
		if msg := item.Validate(); msg != "" {
			return msg
		}
	}
	if r.CouponCode != "" {
		if len(r.CouponCode) < 8 || len(r.CouponCode) > 10 {
			return "couponCode must be between 8 and 10 characters"
		}
	}
	return ""
}

// OrderLine is a product + quantity in the response.
type OrderLine struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// Order is the response for POST /order.
type Order struct {
	ID        string      `json:"id"`
	Total     float64     `json:"total"`
	Discounts float64     `json:"discounts"`
	Items     []OrderLine `json:"items"`
	Products  []Product   `json:"products"`
}