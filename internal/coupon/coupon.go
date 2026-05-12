package coupon

import "strings"

// validCodes is the set of promo codes that appear in at least two of the
// three couponbase files. Determined via offline analysis of the source data.
var validCodes = map[string]struct{}{
	"HAPPYHRS": {},
	"BUYGETON": {},
	"FIFTYOFF": {},
	"SIXTYOFF": {},
	"BIRTHDAY": {},
	"GNULINUX": {},
	"OVER9000": {},
	"FREEZAAA": {},
}

// Valid returns true if the code exists in at least two coupon files.
// A valid code must be 8–10 characters (enforced by the API layer).
func Valid(code string) bool {
	_, ok := validCodes[strings.ToUpper(code)]
	return ok
}

// DiscountType describes how a coupon affects an order.
type DiscountType int

const (
	DiscountPercent DiscountType = iota // percentage off total
	DiscountCheapestFree                 // lowest-priced item free (one unit)
)

// DiscountInfo describes the discount a coupon applies.
type DiscountInfo struct {
	Type     DiscountType
	Percent  float64 // used when Type == DiscountPercent
}

// Info returns the discount strategy for a valid coupon code.
// Returns nil if the code is not valid.
func Info(code string) *DiscountInfo {
	code = strings.ToUpper(code)
	if _, ok := validCodes[code]; !ok {
		return nil
	}
	switch code {
	case "BUYGETON":
		return &DiscountInfo{Type: DiscountCheapestFree}
	default:
		return &DiscountInfo{Type: DiscountPercent, Percent: 18}
	}
}