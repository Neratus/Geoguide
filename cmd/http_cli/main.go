package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	var (
		apiBase     string
		interactive bool
	)
	flag.StringVar(&apiBase, "api", "http://localhost:8080", "base URL of a running GeoGuide API")
	flag.BoolVar(&interactive, "i", true, "interactive mode (default). Use -i=false with positional args for one-shot command")
	flag.Parse()

	client := NewHTTPCLI(apiBase)

	if flag.NArg() > 0 {
		client.RunCommandLine(flag.Args())
		return
	}
	if interactive {
		client.RunInteractive()
		return
	}
	fmt.Println("Nothing to do. Pass a command, or use -i for interactive mode.")
	os.Exit(1)
}

type HTTPCLI struct {
	baseURL string
	token   string
	userID  string
	http    *http.Client
}

func NewHTTPCLI(baseURL string) *HTTPCLI {
	return &HTTPCLI{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *HTTPCLI) do(method, path string, query url.Values, body any) (int, []byte, http.Header, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, nil, err
		}
		reader = bytes.NewReader(b)
	}
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequest(method, u, reader)
	if err != nil {
		return 0, nil, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, resp.Header, err
	}
	return resp.StatusCode, data, resp.Header, nil
}

func (c *HTTPCLI) req(method, path string, query url.Values, body any) ([]byte, http.Header, error) {
	status, data, hdr, err := c.do(method, path, query, body)
	if err != nil {
		return nil, nil, err
	}
	if status >= 400 {
		return data, hdr, fmt.Errorf("HTTP %d: %s", status, strings.TrimSpace(string(data)))
	}
	return data, hdr, nil
}

func (c *HTTPCLI) doMultipart(path, fileField, filePath string) (int, []byte, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	f, err := os.Open(filePath)
	if err != nil {
		return 0, nil, err
	}
	defer f.Close()
	fw, err := w.CreateFormFile(fileField, filepath.Base(filePath))
	if err != nil {
		return 0, nil, err
	}
	if _, err := io.Copy(fw, f); err != nil {
		return 0, nil, err
	}
	w.Close()

	req, err := http.NewRequest("POST", c.baseURL+path, &buf)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data, nil
}

func prettyJSON(data []byte) string {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return string(data)
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return string(data)
	}
	return string(b)
}

func (c *HTTPCLI) printJSON(data []byte) {
	if len(bytes.TrimSpace(data)) == 0 {
		fmt.Println("OK")
		return
	}
	fmt.Println(prettyJSON(data))
}

func (c *HTTPCLI) getJSON(path string, query url.Values) {
	data, _, err := c.req("GET", path, query, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	c.printJSON(data)
}

func (c *HTTPCLI) sendJSON(method, path string, body any) {
	data, _, err := c.req(method, path, nil, body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	c.printJSON(data)
}

func (c *HTTPCLI) RunInteractive() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("GeoGuide HTTP CLI → %s\n", c.baseURL)
	fmt.Println("Type 'help' for commands, 'state' to inspect session, 'exit' to quit.")
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "input error: %v\n", err)
			}
			return
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		cmd, args := parseCommandLine(line)
		if cmd == "exit" || cmd == "quit" {
			return
		}
		c.dispatch(cmd, args)
	}
}

func (c *HTTPCLI) RunCommandLine(args []string) {
	if len(args) == 0 {
		fmt.Println("No command provided. Use -i for interactive mode or 'help'.")
		os.Exit(1)
	}
	c.dispatch(args[0], args[1:])
}

func (c *HTTPCLI) dispatch(cmd string, args []string) {
	switch cmd {
	case "":
		return
	case "help":
		printHTTPHelp()

	case "state", "whoami":
		c.cmdState()
	case "set_token":
		c.cmdSetToken(args)
	case "set_user":
		c.cmdSetUser(args)
	case "clear":
		c.cmdClear()

	case "register":
		c.cmdRegister(args)
	case "verify":
		c.cmdVerify(args)
	case "login":
		c.cmdLogin(args)
	case "verify_2fa":
		c.cmdVerify2FA(args)
	case "logout":
		c.cmdLogout()
	case "profile":
		c.cmdProfile()
	case "update_profile":
		c.cmdUpdateProfile(args)
	case "upload_avatar":
		c.cmdUploadAvatar(args)
	case "enable_2fa":
		c.cmdEnable2FA(args)
	case "moderate_user":
		c.cmdModerateUser(args)

	case "countries":
		c.cmdCountries(args)
	case "country":
		c.cmdCountry(args)
	case "cities":
		c.cmdCities(args)
	case "city":
		c.cmdCity(args)
	case "search_cities":
		c.cmdSearchCities(args)
	case "districts":
		c.cmdDistricts(args)
	case "transport_nodes":
		c.cmdTransportNodes(args)
	case "place":
		c.cmdPlace(args)
	case "places_by_city":
		c.cmdPlacesByCity(args)
	case "places_by_category":
		c.cmdPlacesByCategory(args)
	case "holidays":
		c.cmdHolidays(args)

	case "create_trip":
		c.cmdCreateTrip(args)
	case "get_trip":
		c.cmdGetTrip(args)
	case "user_trips":
		c.cmdUserTrips()
	case "add_place_to_trip":
		c.cmdAddPlaceToTrip(args)
	case "remove_place_from_trip":
		c.cmdRemovePlaceFromTrip(args)
	case "update_trip":
		c.cmdUpdateTrip(args)
	case "delete_trip":
		c.cmdDeleteTrip(args)

	case "create_review":
		c.cmdCreateReview(args)
	case "reviews_by_place":
		c.cmdReviewsByPlace(args)
	case "user_reviews":
		c.cmdUserReviews(args)
	case "moderate_review":
		c.cmdModerateReview(args)
	case "delete_review":
		c.cmdDeleteReview(args)
	case "pending_reviews":
		c.cmdPendingReviews(args)

	case "get_page":
		c.cmdGetPage(args)

	case "add_favourite":
		c.cmdAddFavourite(args)
	case "remove_favourite":
		c.cmdRemoveFavourite(args)
	case "favourites":
		c.cmdFavourites(args)

	case "users":
		c.cmdUsers(args)

	case "report":
		c.cmdReport(args)

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
	}
}

func (c *HTTPCLI) cmdState() {
	fmt.Printf("base_url: %s\n", c.baseURL)
	if c.token == "" {
		fmt.Println("token:    (none)")
	} else {
		fmt.Printf("token:    %s\n", c.token)
	}
	if c.userID == "" {
		fmt.Println("user_id:  (none)")
	} else {
		fmt.Printf("user_id:  %s\n", c.userID)
	}
}

func (c *HTTPCLI) cmdSetToken(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: set_token <jwt>")
		return
	}
	c.token = args[0]
	fmt.Println("token set")
}

func (c *HTTPCLI) cmdSetUser(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: set_user <user_id>")
		return
	}
	c.userID = args[0]
	fmt.Println("user_id set")
}

func (c *HTTPCLI) cmdClear() {
	c.token = ""
	c.userID = ""
	fmt.Println("session cleared")
}

func (c *HTTPCLI) cmdRegister(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: register <username> <email> <password> [phone]")
		return
	}
	body := map[string]any{
		"username": args[0],
		"email":    args[1],
		"password": args[2],
	}
	if len(args) >= 4 {
		body["phone"] = args[3]
	}
	data, _, err := c.req("POST", "/api/v1/auth/register", nil, body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	c.printJSON(data)

	var resp struct {
		ID string `json:"id"`
	}
	if json.Unmarshal(data, &resp) == nil && resp.ID != "" {
		c.userID = resp.ID
		fmt.Printf("→ state.user_id = %s\n", c.userID)
	}
}

func (c *HTTPCLI) cmdVerify(args []string) {
	var userID, contact, code string
	switch len(args) {
	case 3:
		userID, contact, code = args[0], args[1], args[2]
	case 2:
		if c.userID == "" {
			fmt.Println("Usage: verify <contact> <code>  (или set_user <user_id> заранее)")
			return
		}
		userID, contact, code = c.userID, args[0], args[1]
	default:
		fmt.Println("Usage: verify <contact> <code>  OR  verify <user_id> <contact> <code>")
		return
	}
	body := map[string]any{"userId": userID, "contact": contact, "code": code}
	data, _, err := c.req("POST", "/api/v1/auth/verify", nil, body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	c.printJSON(data)
}

func (c *HTTPCLI) cmdLogin(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: login <username> <password>")
		return
	}
	body := map[string]any{"username": args[0], "password": args[1]}
	data, _, err := c.req("POST", "/api/v1/auth/login", nil, body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	var resp struct {
		ID                string `json:"id"`
		Username          string `json:"username"`
		Email             string `json:"email"`
		Token             string `json:"token"`
		RequiresTwoFactor bool   `json:"requiresTwoFactor"`
		ChallengeID       string `json:"challengeId"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		c.printJSON(data)
		return
	}
	if resp.RequiresTwoFactor {
		fmt.Printf("2FA required. challenge_id=%s\n→ verify_2fa %s <code>\n", resp.ChallengeID, resp.ChallengeID)
		return
	}
	c.token = resp.Token
	c.userID = resp.ID
	fmt.Printf("Login OK\n  user_id:  %s\n  username: %s\n  email:    %s\n  token:    %s\n→ state updated\n",
		resp.ID, resp.Username, resp.Email, resp.Token)
}

func (c *HTTPCLI) cmdVerify2FA(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: verify_2fa <challenge_id> <code>")
		return
	}
	body := map[string]any{"challengeId": args[0], "code": args[1]}
	data, _, err := c.req("POST", "/api/v1/auth/2fa/verify", nil, body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	var resp struct {
		Token string `json:"token"`
	}
	if json.Unmarshal(data, &resp) == nil && resp.Token != "" {
		c.token = resp.Token
		fmt.Printf("2FA OK. token saved to state.\n  token: %s\n", resp.Token)
		return
	}
	c.printJSON(data)
}

func (c *HTTPCLI) cmdLogout() {
	if c.token == "" {
		fmt.Println("Not logged in (no token in state)")
		return
	}
	_, _, err := c.req("POST", "/api/v1/auth/logout", nil, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	c.token = ""
	c.userID = ""
	fmt.Println("Logged out. state cleared.")
}

func (c *HTTPCLI) cmdProfile() {
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	c.getJSON("/api/v1/auth/profile", nil)
}

func (c *HTTPCLI) cmdUpdateProfile(args []string) {
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	if len(args) < 1 {
		fmt.Println("Usage: update_profile field=value [field=value...]")
		fmt.Println("Fields: username, phone, birthDate, countryOfResidence, avatarUrl, favoriteCategories (comma-sep)")
		return
	}
	body := map[string]any{}
	for _, kv := range args {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k, v := parts[0], parts[1]
		switch k {
		case "favoriteCategories", "categories":
			body["favoriteCategories"] = strings.Split(v, ",")
		case "username", "phone", "birthDate", "countryOfResidence", "avatarUrl":
			body[k] = v
		}
	}
	if len(body) == 0 {
		fmt.Println("No recognized fields")
		return
	}
	c.sendJSON("PATCH", "/api/v1/auth/profile", body)
}

func (c *HTTPCLI) cmdUploadAvatar(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: upload_avatar <file_path>")
		return
	}
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	status, data, err := c.doMultipart("/api/v1/auth/avatar", "avatar", args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if status >= 400 {
		fmt.Printf("Error: HTTP %d: %s\n", status, string(data))
		return
	}
	c.printJSON(data)
}

func (c *HTTPCLI) cmdEnable2FA(args []string) {
	if c.token == "" || c.userID == "" {
		fmt.Println("No token/user_id. Run 'login' first.")
		return
	}
	if len(args) < 1 {
		fmt.Println("Usage: enable_2fa <secret> <code>")
		return
	}
	if args[0] == "generate" {
		fmt.Println("TOTP secret generation is local-only — use the local CLI for that.")
		fmt.Println("Alternatively, get a secret from Google Authenticator and pass it via:")
		fmt.Println("  enable_2fa <BASE32_SECRET> <6-digit-code>")
		return
	}
	if len(args) < 2 {
		fmt.Println("Usage: enable_2fa <secret> <code>")
		return
	}
	body := map[string]any{
		"userId": c.userID,
		"secret": args[0],
		"code":   args[1],
	}
	c.sendJSON("POST", "/api/v1/auth/2fa/enable", body)
}

func (c *HTTPCLI) cmdModerateUser(args []string) {
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	if len(args) < 2 {
		fmt.Println("Usage: moderate_user <user_id> <true|false> [reason]")
		return
	}
	block, err := strconv.ParseBool(args[1])
	if err != nil {
		fmt.Println("Second arg must be true|false")
		return
	}
	body := map[string]any{"userId": args[0], "block": block}
	if len(args) >= 3 {
		body["reason"] = args[2]
	}
	c.sendJSON("POST", "/api/v1/moderation/users", body)
}

func (c *HTTPCLI) cmdCountries(args []string) {
	q := url.Values{}
	if len(args) >= 1 {
		q.Set("limit", args[0])
	}
	if len(args) >= 2 {
		q.Set("offset", args[1])
	}
	if len(args) >= 3 {
		q.Set("search", args[2])
	}
	c.getJSON("/api/v1/countries", q)
}

func (c *HTTPCLI) cmdCountry(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: country <id>")
		return
	}
	c.getJSON("/api/v1/countries/"+args[0], nil)
}

func (c *HTTPCLI) cmdCities(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: cities <country_id> [limit] [offset]")
		return
	}
	q := url.Values{}
	if len(args) >= 2 {
		q.Set("limit", args[1])
	}
	if len(args) >= 3 {
		q.Set("offset", args[2])
	}
	c.getJSON("/api/v1/countries/"+args[0]+"/cities", q)
}

func (c *HTTPCLI) cmdCity(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: city <id>")
		return
	}
	c.getJSON("/api/v1/cities/"+args[0], nil)
}

func (c *HTTPCLI) cmdSearchCities(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: search_cities <query> [limit] [offset]")
		return
	}
	q := url.Values{}
	q.Set("query", args[0])
	if len(args) >= 2 {
		q.Set("limit", args[1])
	}
	if len(args) >= 3 {
		q.Set("offset", args[2])
	}
	c.getJSON("/api/v1/cities", q)
}

func (c *HTTPCLI) cmdDistricts(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: districts <city_id> [limit] [offset]")
		return
	}
	q := url.Values{}
	if len(args) >= 2 {
		q.Set("limit", args[1])
	}
	if len(args) >= 3 {
		q.Set("offset", args[2])
	}
	c.getJSON("/api/v1/cities/"+args[0]+"/districts", q)
}

func (c *HTTPCLI) cmdTransportNodes(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: transport_nodes <city_id> [limit] [offset]")
		return
	}
	q := url.Values{}
	if len(args) >= 2 {
		q.Set("limit", args[1])
	}
	if len(args) >= 3 {
		q.Set("offset", args[2])
	}
	c.getJSON("/api/v1/cities/"+args[0]+"/transports", q)
}

func (c *HTTPCLI) cmdPlace(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: place <id>")
		return
	}
	c.getJSON("/api/v1/places/"+args[0], nil)
}

func (c *HTTPCLI) cmdPlacesByCity(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: places_by_city <city_id> [category] [sortBy:rating|name|visits] [limit] [offset]")
		return
	}
	q := url.Values{}
	q.Set("cityId", args[0])
	if len(args) >= 2 && args[1] != "" {
		q.Set("category", args[1])
	}
	if len(args) >= 3 && args[2] != "" {
		q.Set("sortBy", args[2])
	}
	if len(args) >= 4 {
		q.Set("limit", args[3])
	}
	if len(args) >= 5 {
		q.Set("offset", args[4])
	}
	c.getJSON("/api/v1/places", q)
}

func (c *HTTPCLI) cmdPlacesByCategory(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: places_by_category <category> [limit] [offset]")
		return
	}
	q := url.Values{}
	q.Set("category", args[0])
	if len(args) >= 2 {
		q.Set("limit", args[1])
	}
	if len(args) >= 3 {
		q.Set("offset", args[2])
	}
	c.getJSON("/api/v1/places", q)
}

func (c *HTTPCLI) cmdHolidays(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: holidays <YYYY-MM-DD>")
		return
	}
	q := url.Values{}
	q.Set("date", args[0])
	c.getJSON("/api/v1/holidays", q)
}

func (c *HTTPCLI) cmdCreateTrip(args []string) {
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	if len(args) < 4 {
		fmt.Println(`Usage: create_trip <title> <start_date> <end_date> <budget> [notes]`)
		return
	}
	body := map[string]any{
		"title":     args[0],
		"startDate": args[1],
		"endDate":   args[2],
	}
	if b, err := strconv.ParseFloat(args[3], 64); err == nil {
		body["budget"] = b
	}
	if len(args) >= 5 {
		body["notes"] = args[4]
	}
	c.sendJSON("POST", "/api/v1/trips", body)
}

func (c *HTTPCLI) cmdGetTrip(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: get_trip <trip_id>")
		return
	}
	c.getJSON("/api/v1/trips/"+args[0], nil)
}

func (c *HTTPCLI) cmdUserTrips() {
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	c.getJSON("/api/v1/trips", nil)
}

func (c *HTTPCLI) cmdAddPlaceToTrip(args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: add_place_to_trip <trip_id> <place_id> <day_number> <duration_min> [arrival_RFC3339] [notes]")
		return
	}
	body := map[string]any{
		"placeId":   args[1],
		"dayNumber": atoiOr(args[2], 0),
	}
	body["durationMin"] = atoiOr(args[3], 0)
	if len(args) >= 5 && args[4] != "" {
		body["arrivalTime"] = args[4]
	}
	if len(args) >= 6 {
		body["notes"] = args[5]
	}
	c.sendJSON("POST", "/api/v1/trips/"+args[0]+"/places", body)
}

func (c *HTTPCLI) cmdRemovePlaceFromTrip(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: remove_place_from_trip <trip_id> <place_id>")
		return
	}
	q := url.Values{}
	q.Set("placeId", args[1])
	data, _, err := c.req("DELETE", "/api/v1/trips/"+args[0]+"/places", q, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	c.printJSON(data)
}

func (c *HTTPCLI) cmdUpdateTrip(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: update_trip <trip_id> [title] [start_date] [end_date] [budget] [status] [notes]")
		fmt.Println("  Empty string = keep current. Or use key=value pairs: title=... status=... budget=...")
		return
	}
	body := map[string]any{}
	rest := args[1:]

	if len(rest) > 0 && strings.Contains(rest[0], "=") {
		for _, kv := range rest {
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) != 2 {
				continue
			}
			k, v := parts[0], parts[1]
			switch k {
			case "title":
				body["title"] = v
			case "startDate", "start_date":
				body["startDate"] = v
			case "endDate", "end_date":
				body["endDate"] = v
			case "budget":
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					body["budget"] = f
				}
			case "status":
				body["status"] = v
			case "notes":
				body["notes"] = v
			}
		}
	} else {
		if len(rest) > 0 && rest[0] != "" {
			body["title"] = rest[0]
		}
		if len(rest) > 1 && rest[1] != "" {
			body["startDate"] = rest[1]
		}
		if len(rest) > 2 && rest[2] != "" {
			body["endDate"] = rest[2]
		}
		if len(rest) > 3 && rest[3] != "" {
			if f, err := strconv.ParseFloat(rest[3], 64); err == nil {
				body["budget"] = f
			}
		}
		if len(rest) > 4 && rest[4] != "" {
			body["status"] = rest[4]
		}
		if len(rest) > 5 && rest[5] != "" {
			body["notes"] = rest[5]
		}
	}
	if len(body) == 0 {
		fmt.Println("Nothing to update.")
		return
	}
	c.sendJSON("PATCH", "/api/v1/trips/"+args[0], body)
}

func (c *HTTPCLI) cmdDeleteTrip(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: delete_trip <trip_id>")
		return
	}
	data, _, err := c.req("DELETE", "/api/v1/trips/"+args[0], nil, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	c.printJSON(data)
}

func (c *HTTPCLI) cmdCreateReview(args []string) {
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	if len(args) < 4 {
		fmt.Println(`Usage: create_review <place_id> <rating:1..5> <comment> <visit_date:YYYY-MM-DD> [image_url]`)
		return
	}
	rating := atoiOr(args[1], 0)
	if rating < 1 || rating > 5 {
		fmt.Println("Rating must be 1..5")
		return
	}
	body := map[string]any{
		"rating":    rating,
		"comment":   args[2],
		"visitDate": args[3],
	}
	if len(args) >= 5 {
		body["imageUrl"] = args[4]
	}
	c.sendJSON("POST", "/api/v1/places/"+args[0]+"/reviews", body)
}

func (c *HTTPCLI) cmdReviewsByPlace(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: reviews_by_place <place_id> [onlyApproved:true|false] [limit] [offset]")
		return
	}
	q := url.Values{}
	if len(args) >= 2 {
		q.Set("onlyApproved", args[1])
	}
	if len(args) >= 3 {
		q.Set("limit", args[2])
	}
	if len(args) >= 4 {
		q.Set("offset", args[3])
	}
	c.getJSON("/api/v1/places/"+args[0]+"/reviews", q)
}

func (c *HTTPCLI) cmdUserReviews(args []string) {
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	q := url.Values{}
	if len(args) >= 1 {
		q.Set("limit", args[0])
	}
	if len(args) >= 2 {
		q.Set("offset", args[1])
	}
	c.getJSON("/api/v1/users/me/reviews", q)
}

func (c *HTTPCLI) cmdModerateReview(args []string) {
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	if len(args) < 2 {
		fmt.Println("Usage: moderate_review <review_id> <true|false> [comment]")
		return
	}
	approved, err := strconv.ParseBool(args[1])
	if err != nil {
		fmt.Println("Second arg must be true|false")
		return
	}
	body := map[string]any{"reviewId": args[0], "approved": approved}
	if len(args) >= 3 {
		body["comment"] = args[2]
	}
	c.sendJSON("POST", "/api/v1/moderation/reviews", body)
}

func (c *HTTPCLI) cmdDeleteReview(args []string) {
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	if len(args) < 1 {
		fmt.Println("Usage: delete_review <review_id>")
		return
	}
	data, _, err := c.req("DELETE", "/api/v1/reviews/"+args[0], nil, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	c.printJSON(data)
}

func (c *HTTPCLI) cmdPendingReviews(args []string) {
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	q := url.Values{}
	if len(args) >= 1 {
		q.Set("limit", args[0])
	}
	if len(args) >= 2 {
		q.Set("offset", args[1])
	}
	if len(args) >= 3 {
		q.Set("isApproved", args[2])
	}
	c.getJSON("/api/v1/moderation/reviews", q)
}

func (c *HTTPCLI) cmdGetPage(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: get_page <id>")
		return
	}
	c.getJSON("/api/v1/files/"+args[0], nil)
}

func (c *HTTPCLI) cmdAddFavourite(args []string) {
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	if len(args) < 1 {
		fmt.Println("Usage: add_favourite <place_id>")
		return
	}
	body := map[string]any{"placeId": args[0]}
	c.sendJSON("POST", "/api/v1/favourites", body)
}

func (c *HTTPCLI) cmdRemoveFavourite(args []string) {
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	if len(args) < 1 {
		fmt.Println("Usage: remove_favourite <place_id>")
		return
	}
	body := map[string]any{"placeId": args[0]}
	data, _, err := c.req("DELETE", "/api/v1/favourites", nil, body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	c.printJSON(data)
}

func (c *HTTPCLI) cmdFavourites(args []string) {
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	q := url.Values{}
	if len(args) >= 1 {
		q.Set("limit", args[0])
	}
	if len(args) >= 2 {
		q.Set("offset", args[1])
	}
	c.getJSON("/api/v1/favourites", q)
}

func (c *HTTPCLI) cmdUsers(args []string) {
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	q := url.Values{}
	if len(args) >= 1 {
		q.Set("limit", args[0])
	}
	if len(args) >= 2 {
		q.Set("offset", args[1])
	}
	if len(args) >= 3 {
		q.Set("search", args[2])
	}
	c.getJSON("/api/v1/moderation/users", q)
}

func (c *HTTPCLI) cmdReport(args []string) {
	if c.token == "" {
		fmt.Println("No token. Run 'login' first.")
		return
	}
	if len(args) < 2 {
		fmt.Println("Usage: report <popular_places|user_activity|trip_statistics> <json|csv> [date_from] [date_to] [limit]")
		return
	}
	body := map[string]any{
		"reportType": args[0],
		"format":     args[1],
	}
	if len(args) >= 3 {
		body["dateFrom"] = args[2]
	}
	if len(args) >= 4 {
		body["dateTo"] = args[3]
	}
	if len(args) >= 5 {
		body["limit"] = atoiOr(args[4], 100)
	}
	status, data, hdr, err := c.do("POST", "/api/v1/analytics/reports", nil, body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if status >= 400 {
		fmt.Printf("Error: HTTP %d: %s\n", status, string(data))
		return
	}
	fmt.Printf("HTTP %d\n", status)
	if loc := hdr.Get("Location"); loc != "" {
		fmt.Printf("Location: %s\n", loc)
	}
	if len(data) > 0 {
		c.printJSON(data)
	}
}

func atoiOr(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}

func parseCommandLine(line string) (string, []string) {
	var args []string
	var current strings.Builder
	inQuote := false
	for _, r := range line {
		switch r {
		case '"':
			inQuote = !inQuote
		case ' ':
			if inQuote {
				current.WriteRune(r)
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	if len(args) == 0 {
		return "", nil
	}
	return args[0], args[1:]
}

func printHTTPHelp() {
	fmt.Print(`GeoGuide HTTP CLI — remote mode (talks to a running GeoGuide API).

  Start:  go run ./cmd/httpcli -api http://localhost:8080
  One-shot: go run ./cmd/httpcli -api http://localhost:8080 countries

  === SESSION STATE ===
  state | whoami               — show base_url, token, user_id
  set_token <jwt>              — set token manually
  set_user  <user_id>          — set user_id manually
  clear                        — reset session state
  help                         — this help
  exit | quit                  — leave

  === AUTH ===
  register <username> <email> <password> [phone]
  verify <contact> <code>                         (uses state.user_id)
  verify <user_id> <contact> <code>
  login <username> <password>                     (saves token + user_id)
  verify_2fa <challenge_id> <code>                (saves token)
  logout                                          (clears token)
  profile
  update_profile field=value [field=value...]
      fields: username, phone, birthDate, countryOfResidence, avatarUrl, favoriteCategories
  upload_avatar <file_path>
  enable_2fa <secret> <code>
  moderate_user <user_id> <true|false> [reason]

  === CATALOG (public) ===
  countries [limit] [offset] [search]
  country <id>
  cities <country_id> [limit] [offset]
  city <id>
  search_cities <query> [limit] [offset]
  districts <city_id> [limit] [offset]
  transport_nodes <city_id> [limit] [offset]
  place <id>
  places_by_city <city_id> [category] [sortBy:rating|name|visits] [limit] [offset]
  places_by_category <category> [limit] [offset]
  holidays <YYYY-MM-DD>

  === PLANNER (auth) ===
  create_trip <title> <start_date> <end_date> <budget> [notes]
  get_trip <trip_id>
  user_trips
  add_place_to_trip <trip_id> <place_id> <day> <duration_min> [arrival_RFC3339] [notes]
  remove_place_from_trip <trip_id> <place_id>
  update_trip <trip_id> [title] [start_date] [end_date] [budget] [status] [notes]
      or: update_trip <trip_id> field=value [field=value...]
  delete_trip <trip_id>

  === REVIEWS ===
  create_review <place_id> <rating:1..5> <comment> <visit_date> [image_url]
  reviews_by_place <place_id> [onlyApproved:true|false] [limit] [offset]
  user_reviews [limit] [offset]
  moderate_review <review_id> <true|false> [comment]
  delete_review <review_id>
  pending_reviews [limit] [offset] [isApproved:true|false]

  === CONTENT ===
  get_page <id>

  === FAVOURITES ===
  add_favourite <place_id>
  remove_favourite <place_id>
  favourites [limit] [offset]

  === ADMIN ===
  users [limit] [offset] [search]

  === ANALYTICS ===
  report <popular_places|user_activity|trip_statistics> <json|csv> [date_from] [date_to] [limit]
`)
}
