package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"
)

type Config struct {
	Username  string
	Password  string
	StatsFile string
	Title     string
	Tags      []string
	Visibility string
}

func main() {
	// Parse command line arguments
	username := flag.String("username", "", "Letterboxd username")
	password := flag.String("password", "", "Letterboxd password")
	statsFile := flag.String("stats", "", "Path to the stats markdown file")
	title := flag.String("title", "My Movie Collection Statistics", "Title for the review")
	tagsStr := flag.String("tags", "statistics,trakt,export", "Comma-separated list of tags")
	visibility := flag.String("visibility", "public", "Visibility setting (public, private, or friends)")
	configFile := flag.String("config", "", "Path to config file (alternative to command line args)")
	flag.Parse()

	var config Config

	// Load from config file if provided
	if *configFile != "" {
		configData, err := os.ReadFile(*configFile)
		if err != nil {
			fmt.Printf("Error reading config file: %v\n", err)
			os.Exit(1)
		}

		err = json.Unmarshal(configData, &config)
		if err != nil {
			fmt.Printf("Error parsing config file: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Use command line arguments
		config = Config{
			Username:   *username,
			Password:   *password,
			StatsFile:  *statsFile,
			Title:      *title,
			Tags:       strings.Split(*tagsStr, ","),
			Visibility: *visibility,
		}
	}

	// Validate required fields
	if config.Username == "" || config.Password == "" {
		fmt.Println("Error: username and password are required")
		flag.Usage()
		os.Exit(1)
	}

	if config.StatsFile == "" {
		fmt.Println("Error: stats file path is required")
		flag.Usage()
		os.Exit(1)
	}

	// Read the stats markdown content
	content, err := os.ReadFile(config.StatsFile)
	if err != nil {
		fmt.Printf("Error reading stats file: %v\n", err)
		os.Exit(1)
	}

	// Create a cookie jar for session handling
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		fmt.Printf("Error creating cookie jar: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	// Login to Letterboxd
	fmt.Println("Logging in to Letterboxd...")
	err = login(client, config.Username, config.Password)
	if err != nil {
		fmt.Printf("Error logging in: %v\n", err)
		os.Exit(1)
	}

	// Create a diary entry for the stats
	fmt.Println("Creating review with statistics...")
	err = createReview(client, string(content), config)
	if err != nil {
		fmt.Printf("Error creating review: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully posted statistics to Letterboxd!")
}

func login(client *http.Client, username, password string) error {
	// First, get the login page to extract the CSRF token
	resp, err := client.Get("https://letterboxd.com/sign-in/")
	if err != nil {
		return fmt.Errorf("error fetching login page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code when fetching login page: %d", resp.StatusCode)
	}

	// Parse the HTML to find the CSRF token
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return fmt.Errorf("error parsing login page HTML: %w", err)
	}

	csrfToken := extractCSRFToken(doc)
	if csrfToken == "" {
		return fmt.Errorf("could not find CSRF token")
	}

	// Now submit the login form
	loginData := url.Values{}
	loginData.Set("__csrf", csrfToken)
	loginData.Set("username", username)
	loginData.Set("password", password)

	req, err := http.NewRequest("POST", "https://letterboxd.com/user/login.do", strings.NewReader(loginData.Encode()))
	if err != nil {
		return fmt.Errorf("error creating login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://letterboxd.com/sign-in/")

	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("error submitting login request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the login was successful
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading login response: %w", err)
	}

	if strings.Contains(string(body), "Invalid username or password") {
		return fmt.Errorf("invalid username or password")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return fmt.Errorf("unexpected status code after login: %d", resp.StatusCode)
	}

	return nil
}

func createReview(client *http.Client, content string, config Config) error {
	// First, we need to create a "film" to review
	// For our statistics, we'll use a placeholder film - let's use "The General" (1926)
	// This is a well-known film that's likely to be in everyone's database
	filmURL := "https://letterboxd.com/film/the-general/"

	// Get the film page to extract the CSRF token
	resp, err := client.Get(filmURL)
	if err != nil {
		return fmt.Errorf("error fetching film page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code when fetching film page: %d", resp.StatusCode)
	}

	// Parse the HTML to find the CSRF token
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return fmt.Errorf("error parsing film page HTML: %w", err)
	}

	csrfToken := extractCSRFToken(doc)
	if csrfToken == "" {
		return fmt.Errorf("could not find CSRF token")
	}

	// Extract film ID from URL
	re := regexp.MustCompile(`\/film\/([^\/]+)`)
	matches := re.FindStringSubmatch(filmURL)
	if len(matches) < 2 {
		return fmt.Errorf("could not extract film ID from URL")
	}
	filmID := matches[1]

	// Now submit the review form
	reviewData := url.Values{}
	reviewData.Set("__csrf", csrfToken)
	reviewData.Set("filmId", filmID)
	reviewData.Set("rating", "") // No rating
	reviewData.Set("review", fmt.Sprintf("# %s\n\n%s", config.Title, content))
	reviewData.Set("tags", strings.Join(config.Tags, ","))
	reviewData.Set("containsSpoilers", "false")
	reviewData.Set("rewatch", "false")
	
	// Format current date as yyyy-MM-dd
	today := time.Now().Format("2006-01-02")
	reviewData.Set("diaryDate", today)
	
	// Set visibility
	switch strings.ToLower(config.Visibility) {
	case "private":
		reviewData.Set("visibility", "10")
	case "friends":
		reviewData.Set("visibility", "30")
	default:
		reviewData.Set("visibility", "20") // public
	}

	req, err := http.NewRequest("POST", "https://letterboxd.com/film/the-general/review/", strings.NewReader(reviewData.Encode()))
	if err != nil {
		return fmt.Errorf("error creating review request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", filmURL)

	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("error submitting review request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code after submitting review: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func extractCSRFToken(n *html.Node) string {
	var csrfToken string
	var findCSRFToken func(*html.Node)

	findCSRFToken = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			var isCSRF bool
			var token string

			for _, attr := range n.Attr {
				if attr.Key == "name" && attr.Val == "__csrf" {
					isCSRF = true
				}
				if attr.Key == "value" {
					token = attr.Val
				}
			}

			if isCSRF && token != "" {
				csrfToken = token
				return
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if csrfToken != "" {
				return
			}
			findCSRFToken(c)
		}
	}

	findCSRFToken(n)
	return csrfToken
}