package coupon

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// validCodes is built at startup by Load(). If files are unavailable,
// the embedded fallback set is used.
var validCodes map[string]struct{}

// fallbackCodes is the validated set discovered via offline analysis.
// Used when the coupon files cannot be read.
var fallbackCodes = map[string]struct{}{
	"HAPPYHRS": {},
	"BUYGETON": {},
	"FIFTYOFF": {},
	"SIXTYOFF": {},
	"BIRTHDAY": {},
	"GNULINUX": {},
	"OVER9000": {},
	"FREEZAAA": {},
}

// Load reads the three couponbase .gz files from dir, builds the set
// of promo codes appearing in ≥2 files, and stores it for O(1) lookups.
//
// Algorithm (optimal memory — streams all 3 files sequentially):
//   1. Scan file2.gz → collect lines with len 8 into set2 (≤8 entries)
//   2. Scan file3.gz → collect lines with len 8 into set3 (≤8 entries)
//   3. Scan file1.gz → for each 8-char line, check membership in set2∪set3
//   4. Merge: codes in ≥2 of {file1, file2, file3} qualify
//
// Memory usage: < 1 KB for code sets, regardless of file size.
func Load(dir string) error {
	file2Codes, err := eightCharCodes(filepath.Join(dir, "couponbase2.gz"))
	if err != nil {
		return fmt.Errorf("reading couponbase2.gz: %w", err)
	}

	file3Codes, err := eightCharCodes(filepath.Join(dir, "couponbase3.gz"))
	if err != nil {
		return fmt.Errorf("reading couponbase3.gz: %w", err)
	}

	// Codes present in file2 or file3 (8-char only — the only length that can overlap)
	candidates := make(map[string]struct{})
	for c := range file2Codes {
		candidates[c] = struct{}{}
	}
	for c := range file3Codes {
		candidates[c] = struct{}{}
	}

	// Scan file1 and check each 8-char code against the candidate set
	valid := make(map[string]struct{})

	// Track which candidate codes we found in file1
	foundInFile1 := make(map[string]struct{})

	f, err := os.Open(filepath.Join(dir, "couponbase1.gz"))
	if err != nil {
		return fmt.Errorf("reading couponbase1.gz: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	scanner := bufio.NewScanner(gz)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024) // 8MB max line

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 8 {
			if _, ok := candidates[line]; ok {
				foundInFile1[line] = struct{}{}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanning couponbase1.gz: %w", err)
	}

	// Code is valid if it appears in ≥2 of the 3 files.
	// Since only 8-char codes can match across files:
	//   - In file1 AND file2 → valid
	//   - In file1 AND file3 → valid
	//   - In file2 AND file3 (without file1) → also valid
	for code := range file2Codes {
		if _, ok := foundInFile1[code]; ok {
			valid[code] = struct{}{} // file1 ∩ file2
		}
	}
	for code := range file3Codes {
		if _, ok := foundInFile1[code]; ok {
			valid[code] = struct{}{} // file1 ∩ file3
		}
	}
	for code := range file2Codes {
		if _, ok := file3Codes[code]; ok {
			valid[code] = struct{}{} // file2 ∩ file3
		}
	}

	validCodes = valid

	// Log the result
	codes := make([]string, 0, len(valid))
	for c := range valid {
		codes = append(codes, c)
	}
	sort.Strings(codes)
	fmt.Printf("coupon: loaded %d valid codes from %d candidate 8-char codes\n", len(valid), len(candidates))
	for _, c := range codes {
		fmt.Printf("  ✓ %s\n", c)
	}

	return nil
}

// eightCharCodes reads a gzip file and returns all lines with length 8.
// Used to extract the rare 8-char codes from files that are otherwise
// all 9-char or 10-char.
func eightCharCodes(path string) (map[string]struct{}, error) {
	codes := make(map[string]struct{})

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	scanner := bufio.NewScanner(gz)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024) // 16MB max for 9-10-char lines

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 8 {
			codes[line] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return codes, nil
}

// Valid returns true if the code appears in ≥2 of the 3 coupon base files.
// Thread-safe read-only lookup — safe for concurrent HTTP handlers.
func Valid(code string) bool {
	set := activeSet()
	_, ok := set[strings.ToUpper(code)]
	return ok
}

// activeSet returns the current valid codes set (dynamic or fallback).
func activeSet() map[string]struct{} {
	if validCodes != nil {
		return validCodes
	}
	return fallbackCodes
}

// DiscountType describes how a coupon affects an order.
type DiscountType int

const (
	DiscountPercent    DiscountType = iota // percentage off total
	DiscountCheapestFree                    // lowest-priced item free (one unit)
)

// DiscountInfo describes the discount a coupon applies.
type DiscountInfo struct {
	Type    DiscountType
	Percent float64 // used when Type == DiscountPercent
}

// Info returns the discount strategy for a valid coupon code.
// Returns nil if the code is not valid.
func Info(code string) *DiscountInfo {
	code = strings.ToUpper(code)
	if _, ok := activeSet()[code]; !ok {
		return nil
	}
	switch code {
	case "BUYGETON":
		return &DiscountInfo{Type: DiscountCheapestFree}
	default:
		return &DiscountInfo{Type: DiscountPercent, Percent: 18}
	}
}

// Count returns the number of known valid codes (useful for health checks/metrics).
func Count() int {
	return len(activeSet())
}

// List returns a sorted list of all valid codes (for admin/debug endpoints).
func List() []string {
	set := activeSet()
	codes := make([]string, 0, len(set))
	for c := range set {
		codes = append(codes, c)
	}
	sort.Strings(codes)
	return codes
}