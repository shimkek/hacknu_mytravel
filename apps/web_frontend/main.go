package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type Accommodation struct {
	ID                 int              `json:"id"`
	Name               string           `json:"name"`
	Latitude           *float64         `json:"latitude"`
	Longitude          *float64         `json:"longitude"`
	Address            *string          `json:"address"`
	Phone              *string          `json:"phone"`
	Email              *string          `json:"email"`
	WebsiteURL         *string          `json:"website_url"`
	ServiceDescription *string          `json:"service_description"`
	RoomCount          *int             `json:"room_count"`
	Capacity           *int             `json:"capacity"`
	PriceRangeMin      *float64         `json:"price_range_min"`
	PriceRangeMax      *float64         `json:"price_range_max"`
	PriceCurrency      string           `json:"price_currency"`
	Rating             *float64         `json:"rating"`
	ReviewCount        int              `json:"review_count"`
	AccommodationType  *string          `json:"accommodation_type"`
	SourceWebsite      string           `json:"source_website"`
	VerificationStatus string           `json:"verification_status"`
	Amenities          *json.RawMessage `json:"amenities"`
	Reviews            *json.RawMessage `json:"reviews"`
}

type AIAnalysis struct {
	Description string  `json:"description"`
	Score       float64 `json:"score"`
	Evaluation  string  `json:"evaluation"`
}

var db *sql.DB
var aiAnalyzerURL string

func init() {
	var err error

	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "mytravel_db")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	aiAnalyzerURL = getEnv("AI_ANALYZER_URL", "http://localhost:8080")

	log.Println("Connected to database successfully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	defer db.Close()

	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/accommodations", accommodationsHandler)
	http.HandleFunc("/api/accommodation/", accommodationHandler)
	http.HandleFunc("/api/ai/description/", aiDescriptionHandler)
	http.HandleFunc("/api/ai/evaluation/", aiEvaluationHandler)

	port := getEnv("PORT", "3000")
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("home").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MyTravel - Accommodations</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css" rel="stylesheet">
    <style>
        .accommodation-card {
            transition: transform 0.2s;
            height: 100%;
        }
        .accommodation-card:hover {
            transform: translateY(-5px);
        }
        .rating-stars {
            color: #ffc107;
        }
        .ai-analysis {
            background-color: #f8f9fa;
            border-left: 4px solid #007bff;
            padding: 15px;
            margin-top: 15px;
        }
        .loading {
            display: none;
        }
        .filter-section {
            background-color: #f8f9fa;
            padding: 20px;
            border-radius: 10px;
            margin-bottom: 30px;
        }
        .price-range {
            color: #28a745;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <nav class="navbar navbar-expand-lg navbar-dark bg-primary">
        <div class="container">
            <a class="navbar-brand" href="/">
                <i class="fas fa-map-marker-alt"></i> MyTravel
            </a>
        </div>
    </nav>

    <div class="container mt-4">
        <div class="row">
            <div class="col-12">
                <h1>Accommodations in Kazakhstan</h1>
                <p class="text-muted">Discover amazing places to stay with AI-powered insights</p>
            </div>
        </div>

        <!-- Filters -->
        <div class="filter-section">
            <div class="row">
                <div class="col-md-3">
                    <label for="sourceFilter" class="form-label">Source</label>
                    <select class="form-select" id="sourceFilter">
                        <option value="">All Sources</option>
                        <option value="2gis">2GIS</option>
                        <option value="google_maps">Google Maps</option>
                        <option value="instagram">Instagram</option>
                        <option value="olx">OLX</option>
                        <option value="yandex">Yandex</option>
                    </select>
                </div>
                <div class="col-md-3">
                    <label for="typeFilter" class="form-label">Type</label>
                    <select class="form-select" id="typeFilter">
                        <option value="">All Types</option>
                        <option value="Hotel">Hotel</option>
                        <option value="Hostel">Hostel</option>
                        <option value="Resort">Resort</option>
                        <option value="Villa">Villa</option>
                        <option value="Camping">Camping</option>
                    </select>
                </div>
                <div class="col-md-3">
                    <label for="ratingFilter" class="form-label">Min Rating</label>
                    <select class="form-select" id="ratingFilter">
                        <option value="">Any Rating</option>
                        <option value="3">3+ Stars</option>
                        <option value="4">4+ Stars</option>
                        <option value="4.5">4.5+ Stars</option>
                    </select>
                </div>
                <div class="col-md-3">
                    <label for="searchInput" class="form-label">Search</label>
                    <input type="text" class="form-control" id="searchInput" placeholder="Search by name or address">
                </div>
            </div>
            <div class="row mt-3">
                <div class="col-12">
                    <button class="btn btn-primary" onclick="loadAccommodations()">
                        <i class="fas fa-search"></i> Apply Filters
                    </button>
                    <button class="btn btn-outline-secondary" onclick="clearFilters()">
                        <i class="fas fa-times"></i> Clear
                    </button>
                </div>
            </div>
        </div>

        <!-- Loading indicator -->
        <div class="text-center loading" id="mainLoading">
            <div class="spinner-border" role="status">
                <span class="visually-hidden">Loading...</span>
            </div>
        </div>

        <!-- Accommodations count and pagination controls -->
        <div class="row mb-3" id="paginationTop" style="display: none;">
            <div class="col-md-6">
                <p class="text-muted" id="accommodationCount"></p>
            </div>
            <div class="col-md-6 text-end">
                <button class="btn btn-outline-primary btn-sm" id="loadMoreBtn" onclick="loadMore()">
                    <i class="fas fa-plus"></i> Load More (100)
                </button>
            </div>
        </div>

        <!-- Accommodations grid -->
        <div class="row" id="accommodationsGrid">
            <!-- Accommodations will be loaded here -->
        </div>

        <!-- Load more button -->
        <div class="row mt-3" id="paginationBottom" style="display: none;">
            <div class="col-12 text-center">
                <button class="btn btn-primary" id="loadMoreBtnBottom" onclick="loadMore()">
                    <i class="fas fa-plus"></i> Load More Accommodations
                </button>
            </div>
        </div>

    </div>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        let currentAccommodations = [];
        let currentOffset = 0;
        const limit = 100;

        document.addEventListener('DOMContentLoaded', function() {
            loadAccommodations();
        });

        async function loadAccommodations() {
            const loading = document.getElementById('mainLoading');
            const grid = document.getElementById('accommodationsGrid');
            const paginationTop = document.getElementById('paginationTop');
            const paginationBottom = document.getElementById('paginationBottom');
            const accommodationCount = document.getElementById('accommodationCount');
            
            loading.style.display = 'block';
            grid.innerHTML = '';
            paginationTop.style.display = 'none';
            paginationBottom.style.display = 'none';
            accommodationCount.innerHTML = '';

            try {
                const params = new URLSearchParams();
                
                const source = document.getElementById('sourceFilter').value;
                const type = document.getElementById('typeFilter').value;
                const rating = document.getElementById('ratingFilter').value;
                const search = document.getElementById('searchInput').value;
                
                if (source) params.append('source_website', source);
                if (type) params.append('accommodation_type', type);
                if (rating) params.append('min_rating', rating);
                if (search) params.append('search', search);
                
                params.append('limit', limit);
                params.append('offset', currentOffset);
                
                const response = await fetch('/api/accommodations?' + params.toString());
                const accommodations = await response.json();
                
                currentAccommodations = accommodations;
                displayAccommodations(accommodations);

                if (accommodations.length > 0) {
                    paginationTop.style.display = 'flex';
                    paginationBottom.style.display = 'flex';
                    accommodationCount.innerHTML = 'Showing ${currentOffset + 1} to ${currentOffset + accommodations.length} accommodations';
                }
            } catch (error) {
                console.error('Error loading accommodations:', error);
                grid.innerHTML = '<div class="col-12"><div class="alert alert-danger">Error loading accommodations. Please try again.</div></div>';
            } finally {
                loading.style.display = 'none';
            }
        }

        function displayAccommodations(accommodations) {
            const grid = document.getElementById('accommodationsGrid');
            
            if (accommodations.length === 0) {
                grid.innerHTML = '<div class="col-12"><div class="alert alert-info">No accommodations found matching your criteria.</div></div>';
                return;
            }

            grid.innerHTML = accommodations.map(acc => {
                let html = '<div class="col-md-6 col-lg-4 mb-4">';
                html += '<div class="card accommodation-card h-100">';
                html += '<div class="card-body">';
                html += '<h5 class="card-title">' + acc.name + '</h5>';
                
                if (acc.address) {
                    html += '<p class="card-text text-muted"><i class="fas fa-map-marker-alt"></i> ' + acc.address + '</p>';
                }
                
                // Rating and Review Count
                html += '<div class="mb-2">';
                if (acc.rating) {
                    html += '<span class="rating-stars">' + generateStars(acc.rating) + '</span>';
                    html += '<span class="text-muted">(' + acc.rating + '/5)</span>';
                    if (acc.review_count > 0) {
                        html += '<small class="text-muted ms-2">' + acc.review_count + ' reviews</small>';
                    }
                } else {
                    html += '<span class="text-muted">No rating</span>';
                }
                html += '</div>';
                
                // Accommodation Type Badge
                if (acc.accommodation_type) {
                    html += '<span class="badge bg-secondary mb-2">' + acc.accommodation_type + '</span>';
                }
                
                // Price Range
                if (acc.price_range_min && acc.price_range_max) {
                    html += '<p class="price-range">';
                    html += '<i class="fas fa-dollar-sign"></i> ';
                    html += acc.price_range_min + ' - ' + acc.price_range_max + ' ' + acc.price_currency;
                    html += '</p>';
                } else if (acc.price_range_min) {
                    html += '<p class="price-range">';
                    html += '<i class="fas fa-dollar-sign"></i> From ' + acc.price_range_min + ' ' + acc.price_currency;
                    html += '</p>';
                }
                
                // Capacity and Room Count
                if (acc.capacity) {
                    html += '<p><i class="fas fa-users"></i> Capacity: ' + acc.capacity + '</p>';
                }
                if (acc.room_count) {
                    html += '<p><i class="fas fa-bed"></i> Rooms: ' + acc.room_count + '</p>';
                }
                
                // Amenities
                if (acc.amenities) {
                    try {
                        const amenities = JSON.parse(acc.amenities);
                        const amenityList = formatAmenities(amenities);
                        if (amenityList.length > 0) {
                            html += '<div class="mb-2">';
                            html += '<small class="text-muted"><i class="fas fa-concierge-bell"></i> Amenities:</small><br>';
                            html += '<small>' + amenityList.join(', ') + '</small>';
                            html += '</div>';
                        }
                    } catch (e) {
                        // Invalid JSON, skip amenities
                    }
                }
                
                // Detailed Reviews Information
                if (acc.reviews) {
                    try {
                        const reviews = JSON.parse(acc.reviews);
                        if (reviews.general_rating > 0) {
                            html += '<div class="mb-2">';
                            html += '<small class="text-muted">';
                            html += '<i class="fas fa-star"></i> General Rating: ' + reviews.general_rating;
                            if (reviews.general_review_count > 0) {
                                html += ' (' + reviews.general_review_count + ' reviews)';
                            }
                            html += '</small>';
                            html += '</div>';
                        }
                    } catch (e) {
                        // Invalid JSON, skip detailed reviews
                    }
                }
                
                // AI Buttons
                html += '<div class="mt-3">';
                html += '<button class="btn btn-primary btn-sm" onclick="generateDescription(' + acc.id + ')">';
                html += '<i class="fas fa-robot"></i> AI Description';
                html += '</button> ';
                html += '<button class="btn btn-success btn-sm" onclick="generateEvaluation(' + acc.id + ')">';
                html += '<i class="fas fa-star"></i> AI Evaluation';
                html += '</button>';
                html += '</div>';
                
                // AI Analysis Container
                html += '<div id="aiAnalysis' + acc.id + '"></div>';
                html += '</div>';
                
                // Card Footer
                html += '<div class="card-footer text-muted">';
                html += '<small>Source: ' + acc.source_website + '</small>';
                html += '</div>';
                html += '</div>';
                html += '</div>';
                
                return html;
            }).join('');
        }

        function formatAmenities(amenities) {
            const amenityNames = {
                'restaurant': 'Restaurant',
                'wifi': 'Wi-Fi',
                'parking': 'Parking',
                'pool': 'Pool',
                'gym': 'Gym',
                'spa': 'Spa',
                'bar': 'Bar',
                'breakfast': 'Breakfast',
                'room_service': 'Room Service',
                'laundry': 'Laundry',
                'ac': 'Air Conditioning',
                'heating': 'Heating',
                'tv': 'TV',
                'minibar': 'Minibar',
                'safe': 'Safe',
                'balcony': 'Balcony',
                'kitchen': 'Kitchen',
                'pets_allowed': 'Pet-Friendly',
                'smoking_allowed': 'Smoking Allowed',
                'disabled_access': 'Wheelchair Accessible',
                'cash_payment': 'Cash Payment',
                'card_payment': 'Card Payment',
                'qr_payment': 'QR Payment',
                'online_payment': 'Online Payment'
            };
            
            const result = [];
            for (const [key, value] of Object.entries(amenities)) {
                if (value === true) {
                    result.push(amenityNames[key] || key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase()));
                }
            }
            return result;
        }

        function generateStars(rating) {
            const fullStars = Math.floor(rating);
            const hasHalfStar = rating % 1 >= 0.5;
            const emptyStars = 5 - fullStars - (hasHalfStar ? 1 : 0);
            
            return '★'.repeat(fullStars) + 
                   (hasHalfStar ? '☆' : '') + 
                   '☆'.repeat(emptyStars);
        }

        async function generateDescription(accommodationId) {
            const analysisDiv = document.getElementById('aiAnalysis' + accommodationId);
            
            analysisDiv.innerHTML = '<div class="ai-analysis"><div class="d-flex align-items-center"><div class="spinner-border spinner-border-sm me-2" role="status"></div><span>Generating AI description...</span></div></div>';

            try {
                const response = await fetch('/api/ai/description/' + accommodationId);
                const result = await response.json();
                
                // Debug: Log the entire response
                console.log('AI Description Response:', result);
                console.log('Analysis field:', result.analysis);
                console.log('Analysis type:', typeof result.analysis);
                
                // Check if analysis field exists and is not undefined
                if (result.analysis && result.analysis !== undefined) {
                    analysisDiv.innerHTML = '<div class="ai-analysis"><h6><i class="fas fa-robot"></i> AI Generated Description</h6><p>' + result.analysis + '</p></div>';
                } else {
                    analysisDiv.innerHTML = '<div class="ai-analysis"><h6><i class="fas fa-robot"></i> AI Generated Description</h6><p>Response received but analysis field is missing or undefined.</p><pre>' + JSON.stringify(result, null, 2) + '</pre></div>';
                }
            } catch (error) {
                console.error('Error generating description:', error);
                analysisDiv.innerHTML = '<div class="ai-analysis"><div class="alert alert-danger">Failed to generate description. Please try again.</div></div>';
            }
        }

        async function generateEvaluation(accommodationId) {
            const analysisDiv = document.getElementById('aiAnalysis' + accommodationId);
            
            analysisDiv.innerHTML = '<div class="ai-analysis"><div class="d-flex align-items-center"><div class="spinner-border spinner-border-sm me-2" role="status"></div><span>Generating AI evaluation...</span></div></div>';

            try {
                const response = await fetch('/api/ai/evaluation/' + accommodationId);
                const result = await response.json();
                
                // Debug: Log the entire response
                console.log('AI Evaluation Response:', result);
                console.log('Analysis field:', result.analysis);
                console.log('Analysis type:', typeof result.analysis);
                
                // Check if analysis field exists and is not undefined
                if (!result.analysis || result.analysis === undefined) {
                    analysisDiv.innerHTML = '<div class="ai-analysis"><h6><i class="fas fa-star"></i> AI Evaluation</h6><p>Response received but analysis field is missing or undefined.</p><pre>' + JSON.stringify(result, null, 2) + '</pre></div>';
                    return;
                }
                
                // Parse the evaluation JSON response
                try {
                    const evaluation = JSON.parse(result.analysis);
                    
                    let html = '<div class="ai-analysis">';
                    html += '<h6><i class="fas fa-star"></i> AI Evaluation</h6>';
                    html += '<div class="mb-2">';
                    html += '<strong>Overall Rating: ' + evaluation['Общий рейтинг'] + '/10</strong>';
                    html += '<span class="badge bg-' + (evaluation['Категория'] === 'горячий' ? 'success' : evaluation['Категория'] === 'тёплый' ? 'warning' : 'secondary') + ' ms-2">' + evaluation['Категория'] + '</span>';
                    html += '<div class="progress mt-1" style="height: 10px;">';
                    html += '<div class="progress-bar" style="width: ' + (evaluation['Общий рейтинг'] * 10) + '%"></div>';
                    html += '</div>';
                    html += '</div>';
                    
                    // Show detailed scores
                    html += '<div class="row">';
                    const criteria = [
                        'Активность в сети',
                        'Полнота данных', 
                        'Популярность',
                        'Потенциал заполняемости',
                        'Соответствие ЦА',
                        'Коммерческий потенциал'
                    ];
                    
                    criteria.forEach(criterion => {
                        if (evaluation[criterion]) {
                            html += '<div class="col-6 mb-1">';
                            html += '<small class="text-muted">' + criterion + ': ' + evaluation[criterion] + '/10</small>';
                            html += '</div>';
                        }
                    });
                    html += '</div>';
                    html += '</div>';
                    
                    analysisDiv.innerHTML = html;
                } catch (parseError) {
                    // If JSON parsing fails, show the raw analysis
                    console.log('JSON parse error:', parseError);
                    analysisDiv.innerHTML = '<div class="ai-analysis"><h6><i class="fas fa-star"></i> AI Evaluation</h6><p>' + result.analysis + '</p></div>';
                }
            } catch (error) {
                console.error('Error generating evaluation:', error);
                analysisDiv.innerHTML = '<div class="ai-analysis"><div class="alert alert-danger">Failed to generate evaluation. Please try again.</div></div>';
            }
        }

        function clearFilters() {
            document.getElementById('sourceFilter').value = '';
            document.getElementById('typeFilter').value = '';
            document.getElementById('ratingFilter').value = '';
            document.getElementById('searchInput').value = '';
            loadAccommodations();
        }

        function loadMore() {
            currentOffset += limit;
            loadAccommodations();
        }

        document.getElementById('searchInput').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                loadAccommodations();
            }
        });
    </script>
</body>
</html>
`))

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, nil)
}

func accommodationsHandler(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, name, latitude, longitude, address, phone, email, website_url, 
		       service_description, room_count, capacity, price_range_min, price_range_max, 
		       price_currency, rating, review_count, accommodation_type, source_website, 
		       verification_status, amenities, reviews
		FROM accommodations 
		WHERE deleted_at IS NULL
	`

	var conditions []string
	var args []interface{}
	// var totalAccommodations int
	argIndex := 1

	// Apply filters
	if source := r.URL.Query().Get("source_website"); source != "" {
		conditions = append(conditions, fmt.Sprintf("source_website = $%d", argIndex))
		args = append(args, source)
		argIndex++
	}

	if accType := r.URL.Query().Get("accommodation_type"); accType != "" {
		conditions = append(conditions, fmt.Sprintf("accommodation_type = $%d", argIndex))
		args = append(args, accType)
		argIndex++
	}

	if minRating := r.URL.Query().Get("min_rating"); minRating != "" {
		if rating, err := strconv.ParseFloat(minRating, 64); err == nil {
			conditions = append(conditions, fmt.Sprintf("rating >= $%d", argIndex))
			args = append(args, rating)
			argIndex++
		}
	}

	if search := r.URL.Query().Get("search"); search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR address ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	// Parse limit and offset for pagination
	limit := 100 // increased from 50 to show more accommodations
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 500 {
			limit = parsedLimit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += fmt.Sprintf(" ORDER BY rating DESC NULLS LAST, name LIMIT %d OFFSET %d", limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var accommodations []Accommodation
	for rows.Next() {
		var acc Accommodation
		err := rows.Scan(
			&acc.ID, &acc.Name, &acc.Latitude, &acc.Longitude, &acc.Address,
			&acc.Phone, &acc.Email, &acc.WebsiteURL, &acc.ServiceDescription,
			&acc.RoomCount, &acc.Capacity, &acc.PriceRangeMin, &acc.PriceRangeMax,
			&acc.PriceCurrency, &acc.Rating, &acc.ReviewCount, &acc.AccommodationType,
			&acc.SourceWebsite, &acc.VerificationStatus, &acc.Amenities, &acc.Reviews,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		accommodations = append(accommodations, acc)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accommodations)
}

func accommodationHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/accommodation/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid accommodation ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, name, latitude, longitude, address, phone, email, website_url, 
		       service_description, room_count, capacity, price_range_min, price_range_max, 
		       price_currency, rating, review_count, accommodation_type, source_website, 
		       verification_status, amenities, reviews
		FROM accommodations 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var acc Accommodation
	err = db.QueryRow(query, id).Scan(
		&acc.ID, &acc.Name, &acc.Latitude, &acc.Longitude, &acc.Address,
		&acc.Phone, &acc.Email, &acc.WebsiteURL, &acc.ServiceDescription,
		&acc.RoomCount, &acc.Capacity, &acc.PriceRangeMin, &acc.PriceRangeMax,
		&acc.PriceCurrency, &acc.Rating, &acc.ReviewCount, &acc.AccommodationType,
		&acc.SourceWebsite, &acc.VerificationStatus, &acc.Amenities, &acc.Reviews,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Accommodation not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(acc)
}

func aiDescriptionHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/ai/description/")

	// Call AI analyzer service with POST request
	resp, err := http.Post(fmt.Sprintf("%s/api/v1/accommodation/%s/description", aiAnalyzerURL, idStr), "application/json", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	json.NewEncoder(w).Encode(result)
}

func aiEvaluationHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/ai/evaluation/")

	// Call AI analyzer service with POST request
	resp, err := http.Post(fmt.Sprintf("%s/api/v1/accommodation/%s/evaluate", aiAnalyzerURL, idStr), "application/json", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	json.NewEncoder(w).Encode(result)
}
