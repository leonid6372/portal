package structs

type Item struct {
	Item_id     int    `json:"item_id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Photo_path  string `json:"photo_path"`
	Is_active   bool   `json:"is_active" validate:"required"`
}
