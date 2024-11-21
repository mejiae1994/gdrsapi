package steamrating

import (
	"fmt"
	"gdrsapi/pkg/logger"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	ageNeededUrl = "https://store.steampowered.com/agecheck/app/1840080/"
	ageSetUrl    = "https://store.steampowered.com/agecheckset/app/1840080/"
	day          = "23"
	month        = "2"
	year         = "1992"
)

type SteamPageContent struct {
	CapsuleImgUrl    string
	CapsuleDesc      string
	Genres           []string
	Tags             []string
	HighlightImgUrls []string
	AboutGameText    string
	AboutGameLinks   []string
	AboutGameImgUrls []string
}

type SteamScraper struct {
	logger     *logger.AppLogger
	httpClient *http.Client
}

func NewSteamScraper(logger *logger.AppLogger) *SteamScraper {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &SteamScraper{
		logger:     logger,
		httpClient: client,
	}
}

func (s *SteamScraper) VerifySteamAgeCheck(steamUrl string) (io.ReadCloser, error) {
	s.logger.InfoLog.Println("starting age check")
	var sessionID string

	req, err := http.NewRequest("GET", ageNeededUrl, nil)
	if err != nil {
		s.logger.ErrorLog.Println(err.Error())
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.ErrorLog.Println(err.Error())
		return nil, fmt.Errorf("fetching age verification page: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(bodyBytes))
	}

	// Extract session ID from cookies
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "sessionid" {
			sessionID = cookie.Value
			break
		}
	}
	if sessionID == "" {
		s.logger.ErrorLog.Println("session ID cookie not found")
		return nil, fmt.Errorf("session ID cookie not found")
	}

	formData := url.Values{
		"sessionid": {sessionID},
		"ageDay":    {day},
		"ageMonth":  {month},
		"ageYear":   {year},
	}

	// Submit age verification
	verifyReq, err := http.NewRequest("POST", ageSetUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		s.logger.ErrorLog.Println(err.Error())
		return nil, fmt.Errorf("creating verification request: %w", err)
	}
	verifyReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	verifyReq.AddCookie(&http.Cookie{Name: "sessionid", Value: sessionID})

	verifyResp, err := s.httpClient.Do(verifyReq)
	if err != nil || verifyResp.StatusCode != http.StatusOK {
		s.logger.ErrorLog.Println(err.Error())
		return nil, fmt.Errorf("age verification error: %w", err)
	}
	defer verifyResp.Body.Close()

	// Fetch game page
	gameReq, err := http.NewRequest("GET", steamUrl, nil)
	if err != nil {
		s.logger.ErrorLog.Println(err.Error())
		return nil, fmt.Errorf("creating game page request: %w", err)
	}

	for _, cookie := range verifyResp.Cookies() {
		gameReq.AddCookie(cookie)
	}
	gameReq.AddCookie(&http.Cookie{Name: "sessionid", Value: sessionID})

	gameResp, err := s.httpClient.Do(gameReq)
	if err != nil || gameResp.StatusCode != http.StatusOK {
		s.logger.ErrorLog.Println(err.Error())
		return nil, fmt.Errorf("fetching game page error: %w", err)
	}

	return gameResp.Body, nil
}

func (s *SteamScraper) ScrapeSteamPage(steamUrl string) (*SteamPageContent, error) {
	pageContent := &SteamPageContent{}

	res, err := s.httpClient.Get(steamUrl)
	if err != nil {
		s.logger.InfoLog.Println("This is inside ScrapeSteamPage")
		s.logger.ErrorLog.Println(err.Error())
		return nil, fmt.Errorf("HTTP error")
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status=%d, body=%s", res.StatusCode, string(bodyBytes))
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		s.logger.ErrorLog.Printf("Failed to parse the HTML document%s", err)
		return nil, fmt.Errorf("failed to parse the HTML document")
	}

	capsuleSection := doc.Find(".glance_ctn")
	if capsuleSection.Length() == 0 {
		s.logger.InfoLog.Println("Retrying to find the capsule section")

		resBody, err := s.VerifySteamAgeCheck(steamUrl)
		if err != nil {
			return nil, err
		}

		doc, _ = goquery.NewDocumentFromReader(resBody)
		capsuleSection = doc.Find(".glance_ctn")
		if capsuleSection.Length() == 0 {
			return nil, fmt.Errorf("failed to find the capsule section")
		}
		s.logger.InfoLog.Println("found capsule section")
	}

	descriptionNode := capsuleSection.Find(".game_description_snippet")
	if descriptionNode.Length() == 0 {
		s.logger.ErrorLog.Println("Failed to grab capsule description")
		return nil, fmt.Errorf("failed to grab capsule description")
	}
	description := strings.TrimSpace(descriptionNode.Text())

	var tags []string
	capsuleSection.Find(".glance_tags.popular_tags .app_tag").Each(func(i int, s *goquery.Selection) {
		tags = append(tags, strings.TrimSpace(s.Text()))
	})

	if len(tags) == 0 {
		s.logger.ErrorLog.Println("no tags found")
		return nil, fmt.Errorf("no tags found")
	}

	//extract genres
	genreSection := doc.Find("#appDetailsUnderlinedLinks")
	if genreSection.Length() == 0 {
		s.logger.ErrorLog.Println("genre section not found")
		return nil, fmt.Errorf("genre section not found")
	}

	genreNodes := genreSection.Find("#genresAndManufacturer > span:first-of-type a")
	if genreNodes.Length() == 0 {
		s.logger.ErrorLog.Println("genre nodes not found")
		return nil, fmt.Errorf("genre nodes not found")
	}

	var genres []string
	genreNodes.Each(func(i int, s *goquery.Selection) {
		genres = append(genres, strings.TrimSpace(s.Text()))
	})

	//extract capsule img
	var capsuleImgUrl string
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		rel, _ := s.Attr("rel")
		if rel == "image_src" {
			capsuleImgUrl, _ = s.Attr("href")
			return
		}
	})

	if capsuleImgUrl == "" {
		return nil, fmt.Errorf("capsule image URL not found")
	}

	//extract imgUrls
	highlightSection := doc.Find("#highlight_player_area")
	if highlightSection.Length() == 0 {
		s.logger.ErrorLog.Println("highlight section not found")
		return nil, fmt.Errorf("highlight section not found")
	}

	var imageUrls []string
	highlightSection.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && href != "" {
			imageUrls = append(imageUrls, href)
		}
	})

	if len(imageUrls) == 0 {
		s.logger.ErrorLog.Println("no image URLs found")
		return nil, fmt.Errorf("no image URLs found")
	}

	//extract about game content
	aboutGameSection := doc.Find("#game_area_description")
	if aboutGameSection.Length() == 0 {
		s.logger.ErrorLog.Println("about game section not found")
		return nil, fmt.Errorf("about game section not found")
	}
	aboutText := strings.TrimSpace(strings.Replace(aboutGameSection.Text(), "About This Game", "", 1))

	//extract about game section urls
	var imgUrls []string
	var linkUrls []string
	aboutGameSection.Find("img, a").Each(func(i int, s *goquery.Selection) {
		if s.Is("img") {
			if src, exists := s.Attr("src"); exists {
				imgUrls = append(imgUrls, src)
			}
		} else if s.Is("a") {
			if href, exists := s.Attr("href"); exists {
				linkUrls = append(linkUrls, href)
			}
		}
	})

	pageContent.CapsuleDesc = description
	pageContent.Tags = tags[:len(tags)-1]
	pageContent.Genres = genres
	pageContent.HighlightImgUrls = imageUrls
	pageContent.AboutGameText = aboutText
	pageContent.AboutGameImgUrls = imgUrls
	pageContent.AboutGameLinks = linkUrls
	pageContent.CapsuleImgUrl = capsuleImgUrl

	return pageContent, nil
}
