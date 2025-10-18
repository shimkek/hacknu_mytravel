package domain

// - Название объекта
// - Координаты (GPS)
// - Адрес
// - Тип размещения (категория)
// - Контактные данные (телефон, email, соцсети)
// - Сайт/страница в соцсети
// - Описание услуг
// - Количество номеров/мест
// - Ценовой диапазон
// - Фотографии (ссылки или мини-галерея)
// - Отзывы и рейтинги (если есть)
// - Инфраструктура (Wi-Fi, паркинг, кухня и т.д.)
// - Статус проверки (новый, проверен, в разработке)
// - Дата последнего обновления

type Business struct {
	Name           string
	Coordinates    Coordinates
	Address        string
	Category       string
	ContactDetails ContactDetails
	Website        string
	Description    string
	Capacity       int
	PriceRange     string
	Photos         []string
	Reviews        []Review
	Amenities      []string
	Verification   string
	LastUpdated    string
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

type ContactDetails struct {
	Phone   string
	Email   string
	Socials []string
}

type Review struct {
	Author  string
	Rating  int
	Comment string
}

// Region represents a 2GIS region
type Region struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// RegionsResponse represents the API response for regions list
type RegionsResponse struct {
	Meta struct {
		APIVersion string `json:"api_version"`
		Code       int    `json:"code"`
		IssueDate  string `json:"issue_date"`
	} `json:"meta"`
	Result struct {
		Items []Region `json:"items"`
		Total int      `json:"total"`
	} `json:"result"`
}
