package parser

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
)

type SummaryProperty struct {
	PropertyName   string `json:"property_name"`
	PageName       string `json:"page_name"`
	Address        string `json:"address"`
	Description    string `json:"description"`
	ReviewsCount   int    `json:"reviews_count"`
	ReviewsRatings string `json:"reviews_ratings"`
}

// func FetchSummaryHTML(url string) (string, error) {
// 	ctx, cancel := chromedp.NewContext(context.Background())
// 	defer cancel()

// 	var html string
// 	err := chromedp.Run(ctx,
// 		chromedp.Navigate(url),
// 		chromedp.Sleep(5*time.Second), // let JS load
// 		chromedp.OuterHTML("html", &html),
// 	)
// 	if err != nil {
// 		return "", fmt.Errorf("chromedp fetch: %v", err)
// 	}

// 	os.WriteFile("summary_page.html", []byte(html), 0644)
// 	return html, nil
// }

func FetchSummary(url string) ([]SummaryProperty, error) {
	// html, err := FetchSummaryHTML(url)
	// if err != nil {
	// 	return nil, err
	// }

	client := &http.Client{Timeout: 25 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Cookie", "bkng=11UmFuZG9tSVYkc2RlIyh9Yaa29%2F3xUOLbwcLxQQ4VaCqIRhUbwbYxY2Eh8f3%2BmPoEGw26gQSh492KJKIjF8dDq1c9AwoiFcwFgRtCy8%2BheSNzsLAxSERO2ORg5EOPiNG21c9W5tXHfU%2BhZOo%2BMvfqmF8s2ijHVuKqqt2AhC1MSdOwVprVq%2FYx6FYUOGGCilXCg7kG2NMQM6I%3D; bkng_sso_auth=CAIQm8CWywYaZtZwyuLD3hUU0dpCx4i6m/SBwiZ8PxXy86X20TZlYebS8HYRkxMC1xvqhvoUkV6lkHjNE8rCpkr+nJ2WouVKeX99nosa4uAK9s8FW+POi4WIAdP9hOBHHiJtt/jeMtCMVe7KB/nDwg==; pcm_consent=analytical%3Dtrue%26countryCode%3DKZ%26consentId%3D42773782-c8a6-405c-ad04-5df1722935ed%26consentedAt%3D2025-10-18T11%3A06%3A33.651Z%26expiresAt%3D2026-04-16T11%3A06%3A33.651Z%26implicit%3Dtrue%26marketing%3Dtrue%26regionCode%3D71%26regulation%3Dnone%26legacyRegulation%3Dnone; pcm_personalization_disabled=0")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch summary: %v", err)
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	switch resp.Header.Get("Content-Encoding") {
	case "br":
		reader = brotli.NewReader(resp.Body)
	case "gzip":
		gr, err := gzip.NewReader(resp.Body)
		if err == nil {
			defer gr.Close()
			reader = gr
		}
	}

	htmlBytes, _ := io.ReadAll(reader)
	html := string(htmlBytes)
	// os.WriteFile("summary_page.html", htmlBytes, 0644) // 💾 debug

	// Relaxed regex: matches any script tag with apollo data
	re := regexp.MustCompile(`(?i)<script[^>]*data-capla-store-data=["']apollo["'][^>]*>([\s\S]*?)</script>`)
	matches := re.FindAllStringSubmatch(html, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("no apollo store found in summary page")
	}

	var rawJSON string
	for _, m := range matches {
		if strings.Contains(m[1], "ROOT_QUERY") {
			rawJSON = strings.TrimSpace(m[1])
			break
		}
	}
	if rawJSON == "" {
		rawJSON = strings.TrimSpace(matches[0][1]) // fallback
	}

	// os.WriteFile("raw_summary.json", []byte(rawJSON), 0644)

	var apollo map[string]interface{}
	if err := json.Unmarshal([]byte(rawJSON), &apollo); err != nil {
		return nil, fmt.Errorf("parse json: %v", err)
	}

	// find searchQueries key dynamically (it might have parameters)
	rootQuery, ok := apollo["ROOT_QUERY"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("ROOT_QUERY not found in Apollo JSON")
	}

	var searchKey string
	for k := range rootQuery {
		if strings.HasPrefix(k, "searchQueries") {
			searchKey = k
			break
		}
	}
	if searchKey == "" {
		return nil, fmt.Errorf("no searchQueries key found")
	}

	// searchObj, ok := rootQuery[searchKey].(map[string]interface{})
	// if !ok {
	// 	return nil, fmt.Errorf("invalid searchQueries format")
	// }

	// resultsRaw, ok := searchObj["results"].([]interface{})
	// if !ok {
	// 	return nil, fmt.Errorf("results not found or invalid format")
	// }

	// recursively find "results" array anywhere in the JSON tree
	var resultsRaw []interface{}

	var findResults func(interface{})
	findResults = func(node interface{}) {
		if resultsRaw != nil {
			return // already found
		}
		switch v := node.(type) {
		case map[string]interface{}:
			for key, val := range v {
				if key == "results" {
					if arr, ok := val.([]interface{}); ok {
						resultsRaw = arr
						return
					}
				}
				findResults(val)
			}
		case []interface{}:
			for _, item := range v {
				findResults(item)
			}
		}
	}

	findResults(rootQuery)

	if resultsRaw == nil {
		fmt.Println("⚠️ Could not find 'results' key. Top-level keys for debugging:")
		for k := range rootQuery {
			fmt.Println(" -", k)
		}
		return nil, fmt.Errorf("results not found or invalid format")
	}

	var results []SummaryProperty
	for _, r := range resultsRaw {
		if prop, ok := r.(map[string]interface{}); ok {
			item := SummaryProperty{}

			if disp, ok := prop["displayName"].(map[string]interface{}); ok {
				if text, ok := disp["text"].(string); ok {
					item.PropertyName = text
				}
			}

			if desc, ok := prop["description"].(map[string]interface{}); ok {
				if text, ok := desc["text"].(string); ok {
					item.Description = text
				}
			}

			if basic, ok := prop["basicPropertyData"].(map[string]interface{}); ok {
				if loc, ok := basic["location"].(map[string]interface{}); ok {
					if addr, ok := loc["address"].(string); ok {
						item.Address = addr
					}
				}
				if pname, ok := basic["pageName"].(string); ok {
					item.PageName = pname
				}
				if reviews, ok := basic["reviews"].(map[string]interface{}); ok {
					if totalScore, ok := reviews["totalScore"].(float64); ok {
						item.ReviewsRatings = fmt.Sprintf("%.1f", totalScore)
					}
					if count, ok := reviews["reviewsCount"].(float64); ok {
						item.ReviewsCount = int(count)
					}
				}
			}

			if item.PropertyName != "" {
				results = append(results, item)
			}
		}
	}

	return results, nil
}
