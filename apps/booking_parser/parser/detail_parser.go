package parser

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Review struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

type DetailedProperty struct {
	PropertyName      string   `json:"property_name"`
	PageName          string   `json:"page_name"`
	URL               string   `json:"url"`
	Latitude          float64  `json:"latitude"`
	Longitude         float64  `json:"longitude"`
	Address           string   `json:"address"`
	AccommodationType string   `json:"accommodation_type"`
	Description       string   `json:"description"`
	Photos            []string `json:"photos"`
	ReviewsRatings    string   `json:"reviews_ratings"`
	ReviewsCount      int      `json:"reviews_count"`
	Reviews           []Review `json:"reviews"`
	Facilities        []string `json:"facilities"`
	InspectionStatus  string   `json:"inspection_status"`
	LastUpdated       string   `json:"last_updated"`
}

func ExtractPropertyDetails(pageURL, defaultDescription, defaultRating string, defaultCount int) (*DetailedProperty, error) {
	client := &http.Client{Timeout: 25 * time.Second}
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var html string

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		gr, err := gzip.NewReader(bytes.NewReader(body))
		if err == nil {
			defer gr.Close()
			unzipped, _ := io.ReadAll(gr)
			html = string(unzipped)
		} else {
			html = string(body)
		}
	case "br":
		br := brotli.NewReader(bytes.NewReader(body))
		unzipped, _ := io.ReadAll(br)
		html = string(unzipped)
	default:
		html = string(body)
	}

	// _ = os.WriteFile("debug_property.html", []byte(html), 0644)

	re := regexp.MustCompile(`<script[^>]*data-capla-store-data="apollo"[^>]*type="application/json"[^>]*>([\s\S]*?)</script>`)
	match := re.FindStringSubmatch(html)
	if len(match) < 2 {
		return nil, fmt.Errorf("! no apollo store found on %s", pageURL)
	}
	rawJSON := strings.TrimSpace(match[1])
	// _ = os.WriteFile("raw_property.json", []byte(rawJSON), 0644)

	var apollo map[string]interface{}
	if err := json.Unmarshal([]byte(rawJSON), &apollo); err != nil {
		return nil, fmt.Errorf("parse json: %v", err)
	}

	result := &DetailedProperty{
		URL:              pageURL,
		Description:      defaultDescription,
		ReviewsRatings:   defaultRating,
		ReviewsCount:     defaultCount,
		InspectionStatus: "New",
		LastUpdated:      time.Now().Format("2006-01-02"),
	}

	for key, val := range apollo {
		switch {
		case strings.HasPrefix(key, "BasicPropertyData:"):
			if bd, ok := val.(map[string]interface{}); ok {
				if n, ok := bd["name"].(string); ok {
					result.PropertyName = n
				}
				if pn, ok := bd["pageName"].(string); ok {
					result.PageName = pn
				}
				if at, ok := bd["accommodationTypeId"].(float64); ok {
					result.AccommodationType = fmt.Sprintf("Type-%d", int(at))
				}
				if loc, ok := bd["location"].(map[string]interface{}); ok {
					if lat, ok := loc["latitude"].(float64); ok {
						result.Latitude = lat
					}
					if lon, ok := loc["longitude"].(float64); ok {
						result.Longitude = lon
					}
					if addr, ok := loc["formattedAddress"].(string); ok {
						result.Address = addr
					}
				}
			}

		case strings.HasPrefix(key, "TextWithTranslationTag:"):
			if desc, ok := val.(map[string]interface{}); ok {
				if txt, ok := desc["text"].(string); ok && result.Description == "" {
					result.Description = txt
				}
			}

		case strings.HasPrefix(key, "AccommodationPhoto:"):
			if ph, ok := val.(map[string]interface{}); ok {
				if res, ok := ph[`resource({"size":"max1024x768"})`].(map[string]interface{}); ok {
					if abs, ok := res["absoluteUrl"].(string); ok {
						cleaned := cleanPhotoURL(abs)
						result.Photos = append(result.Photos, cleaned)
					}
				}
			}

		case strings.HasPrefix(key, "BaseFacility:"):
			if fac, ok := val.(map[string]interface{}); ok {
				if instArr, ok := fac["instances"].([]interface{}); ok {
					for _, inst := range instArr {
						if im, ok := inst.(map[string]interface{}); ok {
							if title, ok := im["title"].(string); ok && title != "" {
								result.Facilities = append(result.Facilities, title)
							}
						}
					}
				}
			}

		case strings.HasPrefix(key, "GenericFacilityHighlight:"):
			if fac, ok := val.(map[string]interface{}); ok {
				if title, ok := fac["title"].(string); ok && title != "" {
					result.Facilities = append(result.Facilities, title)
				}
			}

		case strings.HasPrefix(key, "Property:"):
			if prop, ok := val.(map[string]interface{}); ok {

				// 🏕 Extract accommodation type via reference
				if acctype, ok := prop["accommodationType"].(map[string]interface{}); ok {
					if ref, ok := acctype["__ref"].(string); ok {
						// ref example: PropertyType:{"type":"CAMPING"}
						if propType, ok := apollo[ref].(map[string]interface{}); ok {
							if typ, ok := propType["type"].(string); ok {
								result.AccommodationType = cases.Title(language.English).String(strings.ToLower(typ))
							}
						} else {
							// fallback: parse directly from ref string
							re := regexp.MustCompile(`"type":"([^"]+)"`)
							if match := re.FindStringSubmatch(ref); len(match) > 1 {
								result.AccommodationType = cases.Title(language.English).String(strings.ToLower(match[1]))
							}
						}
					}
				}

				// 🧾 Extract detailed reviews breakdown
				if rev, ok := prop["reviews"].(map[string]interface{}); ok {
					if qs, ok := rev["questions"].([]interface{}); ok {
						for _, q := range qs {
							if qm, ok := q.(map[string]interface{}); ok {
								var r Review
								if name, ok := qm["name"].(string); ok {
									r.Name = name
								}
								if sc, ok := qm["score"].(float64); ok {
									r.Score = sc
								}
								if r.Name != "" {
									result.Reviews = append(result.Reviews, r)
								}
							}
						}
					}
				}
			}

		}
	}

	return result, nil
}

func cleanPhotoURL(url string) string {
	url = strings.ReplaceAll(url, `\u0026`, "&")
	url = strings.TrimSuffix(url, "&o=")
	return strings.TrimSpace(url)
}
