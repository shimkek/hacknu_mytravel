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

// type Business struct {
// 	Name           string
// 	Coordinates    Coordinates
// 	Address        string
// 	Category       string
// 	ContactDetails ContactDetails
// 	Website        string
// 	Description    string
// 	Capacity       int
// 	PriceRange     string
// 	Photos         []string
// 	Reviews        []Review
// 	Amenities      []string
// 	Verification   string
// 	LastUpdated    string
// }

// type Coordinates struct {
// 	Latitude  float64
// 	Longitude float64
// }

// type ContactDetails struct {
// 	Phone   string
// 	Email   string
// 	Socials []string
// }

// type Review struct {
// 	Author  string
// 	Rating  int
// 	Comment string
// }

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

// BusinessByIdResponse represents the API response for business details by ID
type BusinessByIdResponse struct {
	Meta struct {
		APIVersion string `json:"api_version"`
		Code       int    `json:"code"`
		IssueDate  string `json:"issue_date"`
	} `json:"meta"`
	Result struct {
		Items []BusinessDetail `json:"items"`
		Total int              `json:"total"`
	} `json:"result"`
}

// BusinessDetail represents detailed business information from 2GIS API
type BusinessDetail struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Caption         string           `json:"caption"`
	FullName        string           `json:"full_name"`
	FullAddressName string           `json:"full_address_name"`
	AddressName     string           `json:"address_name"`
	Address         Address          `json:"address"`
	Point           Point            `json:"point"`
	PurposeName     string           `json:"purpose_name"`
	Type            string           `json:"type"`
	AttributeGroups []AttributeGroup `json:"attribute_groups"`
	Rubrics         []Rubric         `json:"rubrics"`
	Schedule        Schedule         `json:"schedule"`
	Reviews         ReviewInfo       `json:"reviews"`
	Dates           Dates            `json:"dates"`
	Flags           Flags            `json:"flags"`
	Links           Links            `json:"links"`
	Statistics      Statistics       `json:"statistics"`
	Stat            Stat             `json:"stat"`
}

// Address represents the address information
type Address struct {
	BuildingID string             `json:"building_id"`
	Components []AddressComponent `json:"components"`
}

// AddressComponent represents individual components of an address
type AddressComponent struct {
	Number   string `json:"number"`
	Street   string `json:"street"`
	StreetID string `json:"street_id"`
	Type     string `json:"type"`
}

// Point represents geographical coordinates
type Point struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// AttributeGroup represents grouped attributes/amenities
type AttributeGroup struct {
	Name       string      `json:"name"`
	IconURL    string      `json:"icon_url"`
	IsContext  bool        `json:"is_context"`
	IsPrimary  bool        `json:"is_primary"`
	Attributes []Attribute `json:"attributes"`
	RubricIDs  []string    `json:"rubric_ids"`
}

// Attribute represents individual business attribute/amenity
type Attribute struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

// Rubric represents business category/rubric
type Rubric struct {
	ID       string `json:"id"`
	Alias    string `json:"alias"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	ParentID string `json:"parent_id"`
	ShortID  int    `json:"short_id"`
}

// Schedule represents business working hours
type Schedule struct {
	Mon *DaySchedule `json:"Mon,omitempty"`
	Tue *DaySchedule `json:"Tue,omitempty"`
	Wed *DaySchedule `json:"Wed,omitempty"`
	Thu *DaySchedule `json:"Thu,omitempty"`
	Fri *DaySchedule `json:"Fri,omitempty"`
	Sat *DaySchedule `json:"Sat,omitempty"`
	Sun *DaySchedule `json:"Sun,omitempty"`
}

// DaySchedule represents working hours for a specific day
type DaySchedule struct {
	WorkingHours []WorkingHour `json:"working_hours"`
}

// WorkingHour represents a time period when business is open
type WorkingHour struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// ReviewInfo represents review and rating information
type ReviewInfo struct {
	GeneralRating               float64      `json:"general_rating"`
	GeneralReviewCount          int          `json:"general_review_count"`
	GeneralReviewCountWithStars int          `json:"general_review_count_with_stars"`
	OrgRating                   float64      `json:"org_rating"`
	OrgReviewCount              int          `json:"org_review_count"`
	OrgReviewCountWithStars     int          `json:"org_review_count_with_stars"`
	IsReviewable                bool         `json:"is_reviewable"`
	IsReviewableOnFlamp         bool         `json:"is_reviewable_on_flamp"`
	Items                       []ReviewItem `json:"items"`
}

// ReviewItem represents review platform information
type ReviewItem struct {
	Tag          string `json:"tag"`
	IsReviewable bool   `json:"is_reviewable"`
}

// Dates represents important dates for the business
type Dates struct {
	UpdatedAt string `json:"updated_at"`
}

// Flags represents various boolean flags for the business
type Flags struct {
	Photos bool `json:"photos"`
}

// Links represents related links and references
type Links struct {
	Branches  BranchesInfo `json:"branches"`
	Entrances []Entrance   `json:"entrances"`
}

// BranchesInfo represents branch information
type BranchesInfo struct {
	Count int `json:"count"`
}

// Entrance represents building entrance information
type Entrance struct {
	ID             string   `json:"id"`
	IsPrimary      bool     `json:"is_primary"`
	IsVisibleOnMap bool     `json:"is_visible_on_map"`
	Geometry       Geometry `json:"geometry"`
}

// Geometry represents geographical geometry data
type Geometry struct {
	Points  []string `json:"points"`
	Vectors []string `json:"vectors"`
	Normals []string `json:"normals"`
}

// Statistics represents business statistics
type Statistics struct {
	Area float64 `json:"area"`
}

// Stat represents additional statistics
type Stat struct {
	Adst int64 `json:"adst"`
}

// BusinessSearchResponse represents the API response for business search by keywords
type BusinessSearchResponse struct {
	Meta struct {
		APIVersion string `json:"api_version"`
		Code       int    `json:"code"`
		IssueDate  string `json:"issue_date"`
	} `json:"meta"`
	Result struct {
		Items []BusinessSearchItem `json:"items"`
		Total int                  `json:"total"`
	} `json:"result"`
}

// BusinessSearchItem represents a business item from search results
type BusinessSearchItem struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	FullName    string   `json:"full_name"`
	AddressName string   `json:"address_name"`
	PurposeName string   `json:"purpose_name"`
	Type        string   `json:"type"`
	Rubrics     []Rubric `json:"rubrics"`
}

// RubricData represents collected rubric information
type RubricData struct {
	ID       string `json:"id"`
	Alias    string `json:"alias"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	ParentID string `json:"parent_id"`
	ShortID  int    `json:"short_id"`
	Count    int    `json:"count"` // How many times this rubric appeared
}
