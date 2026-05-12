package model

// Product represents a menu item in the catalog.
type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
	Image    Image   `json:"image"`
}

// Image holds responsive image URLs for a product.
type Image struct {
	Thumbnail string `json:"thumbnail"`
	Mobile    string `json:"mobile"`
	Tablet    string `json:"tablet"`
	Desktop   string `json:"desktop"`
}

// Catalog returns the in-memory product catalog.
// IDs are strings matching the spec examples.
func Catalog() []Product {
	base := "https://orderfoodonline.deno.dev/public/images"
	return []Product{
		{
			ID:       "10",
			Name:     "Chicken Waffle",
			Price:    13.30,
			Category: "Waffle",
			Image: Image{
				Thumbnail: base + "/image-waffle-thumbnail.jpg",
				Mobile:    base + "/image-waffle-mobile.jpg",
				Tablet:    base + "/image-waffle-tablet.jpg",
				Desktop:   base + "/image-waffle-desktop.jpg",
			},
		},
		{
			ID:       "11",
			Name:     "Belgian Waffle",
			Price:    9.50,
			Category: "Waffle",
			Image: Image{
				Thumbnail: base + "/image-belgian-waffle-thumbnail.jpg",
				Mobile:    base + "/image-belgian-waffle-mobile.jpg",
				Tablet:    base + "/image-belgian-waffle-tablet.jpg",
				Desktop:   base + "/image-belgian-waffle-desktop.jpg",
			},
		},
		{
			ID:       "12",
			Name:     "Strawberry Waffle",
			Price:    11.00,
			Category: "Waffle",
			Image: Image{
				Thumbnail: base + "/image-strawberry-waffle-thumbnail.jpg",
				Mobile:    base + "/image-strawberry-waffle-mobile.jpg",
				Tablet:    base + "/image-strawberry-waffle-tablet.jpg",
				Desktop:   base + "/image-strawberry-waffle-desktop.jpg",
			},
		},
		{
			ID:       "20",
			Name:     "Maple Syrup",
			Price:    3.50,
			Category: "Topping",
			Image: Image{
				Thumbnail: base + "/image-maple-syrup-thumbnail.jpg",
				Mobile:    base + "/image-maple-syrup-mobile.jpg",
				Tablet:    base + "/image-maple-syrup-tablet.jpg",
				Desktop:   base + "/image-maple-syrup-desktop.jpg",
			},
		},
		{
			ID:       "21",
			Name:     "Whipped Cream",
			Price:    2.00,
			Category: "Topping",
			Image: Image{
				Thumbnail: base + "/image-whipped-cream-thumbnail.jpg",
				Mobile:    base + "/image-whipped-cream-mobile.jpg",
				Tablet:    base + "/image-whipped-cream-tablet.jpg",
				Desktop:   base + "/image-whipped-cream-desktop.jpg",
			},
		},
		{
			ID:       "22",
			Name:     "Chocolate Sauce",
			Price:    2.50,
			Category: "Topping",
			Image: Image{
				Thumbnail: base + "/image-chocolate-sauce-thumbnail.jpg",
				Mobile:    base + "/image-chocolate-sauce-mobile.jpg",
				Tablet:    base + "/image-chocolate-sauce-tablet.jpg",
				Desktop:   base + "/image-chocolate-sauce-desktop.jpg",
			},
		},
		{
			ID:       "30",
			Name:     "Fresh Orange Juice",
			Price:    4.50,
			Category: "Drink",
			Image: Image{
				Thumbnail: base + "/image-orange-juice-thumbnail.jpg",
				Mobile:    base + "/image-orange-juice-mobile.jpg",
				Tablet:    base + "/image-orange-juice-tablet.jpg",
				Desktop:   base + "/image-orange-juice-desktop.jpg",
			},
		},
		{
			ID:       "31",
			Name:     "Iced Coffee",
			Price:    5.00,
			Category: "Drink",
			Image: Image{
				Thumbnail: base + "/image-iced-coffee-thumbnail.jpg",
				Mobile:    base + "/image-iced-coffee-mobile.jpg",
				Tablet:    base + "/image-iced-coffee-tablet.jpg",
				Desktop:   base + "/image-iced-coffee-desktop.jpg",
			},
		},
	}
}

// Lookup returns a product by ID, or nil if not found.
func Lookup(id string) *Product {
	for _, p := range Catalog() {
		if p.ID == id {
			return &p
		}
	}
	return nil
}