package linkedin

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/GLADIUS240/scrapedin/internal/config"
	"github.com/go-rod/rod"
	"go.uber.org/zap"
)

func BuildPeopleSearchURL(keyword, location string, page int) string {
	params := url.Values{}
	params.Set("keywords", keyword)
	params.Set("origin", "GLOBAL_SEARCH_HEADER")

	if location != "" {
		params.Set("geoUrn", location)
	}

	if page > 1 {
		params.Set("page", fmt.Sprint(page))
	}

	return "https://www.linkedin.com/search/results/people/?" + params.Encode()
}

func ExtractProfileURLs(page *rod.Page) []string {
	anchors := page.MustElements(`a.app-aware-link[href*="/in/"]`)
	unique := make(map[string]struct{})

	for _, a := range anchors {
		href := a.MustAttribute("href")
		if href == nil {
			continue
		}

		url := strings.Split(*href, "?")[0]
		if strings.Contains(url, "/in/") {
			unique[url] = struct{}{}
		}
	}

	results := make([]string, 0, len(unique))
	for u := range unique {
		results = append(results, u)
	}

	return results
}

// internal/linkedin/search.go
func RunPeopleSearch(
	page *rod.Page,
	searchCfg config.SearchConfig,
	logger *zap.Logger,
	onProfiles func(string) error,
) error {

	for _, keyword := range searchCfg.Keywords {
		for _, location := range searchCfg.Locations {

			logger.Info("starting people search",
				zap.String("keyword", keyword),
				zap.String("location", location),
			)

			for p := 1; p <= searchCfg.MaxPages; p++ {
				url := BuildPeopleSearchURL(keyword, location, p)

				page.MustNavigate(url)
				page.MustWaitLoad()

				time.Sleep(time.Duration(searchCfg.PageDelayS) * time.Second)

				profiles := ExtractProfileURLs(page)
				if len(profiles) == 0 {
					break
				}

				for _, profile := range profiles {
					if err := onProfiles(profile); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func HandleReturningUser(page *rod.Page) {
	btn := page.MustElement("button.member-profile__details")
	btn.MustScrollIntoView()
	time.Sleep(500 * time.Millisecond)
	btn.MustClick()
	page.MustWaitLoad()
}
