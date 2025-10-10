package dto

type GenerateCardRequest struct {
	PhotoURL         string `json:"photo_url" validate:"required,url"`
	ShortDescription string `json:"short_description" validate:"required"`
}

type GenerateCardResponse struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Image       string   `json:"image"`
}

type CardHistoryResponse struct {
	Cards []CardInfo `json:"cards"`
}

type CardInfo struct {
	ID               string   `json:"id"`
	PhotoURL         string   `json:"photo_url"`
	ShortDescription string   `json:"short_description"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Tags             []string `json:"tags"`
	Image            string   `json:"image"`
	CreatedAt        string   `json:"created_at"`
}

type CardDetailResponse struct {
	ID               string   `json:"id"`
	PhotoURL         string   `json:"photo_url"`
	ShortDescription string   `json:"short_description"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Tags             []string `json:"tags"`
	Image            string   `json:"image"`
	CreatedAt        string   `json:"created_at"`
}
