package basicauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/auth0-restapi-samples/config"

	"github.com/golang-jwt/jwt/v5"
)

// handleAuthResponse accepts the raw bytes and performs the logic
func handleAuthResponse(rawData []byte) {
	// Unmarshal into a map to handle the JSON dynamically
	var data map[string]interface{}
	if err := json.Unmarshal(rawData, &data); err != nil {
		fmt.Println("Error: Response was not valid JSON")
		return
	}

	// Pretty Print the whole response
	fmt.Println("=== Full JSON Response ===")
	pretty, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(pretty))

	// Extract and Parse JWT if it exists
	if tokenStr, ok := data["access_token"].(string); ok {
		parseAndPrintJWT(tokenStr)
	}
}

func parseAndPrintJWT(tokenStr string) {
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		fmt.Printf("JWT Parse Error: %v\n", err)
		return
	}

	fmt.Println("\n=== JWT Claims (jwt.io style) ===")
	claimsJSON, _ := json.MarshalIndent(token.Claims, "", "  ")
	fmt.Println(string(claimsJSON))
}

func BasicAuthExample(config config.Config) {

	fmt.Println("Config:", config)

	url := "https://" + config.Auth.Domain + "/oauth/token"

	payload := strings.NewReader("{" + "\"client_id\":\"" + config.Auth.M2M.ClientID + "\",\"client_secret\":\"" + config.Auth.M2M.ClientSecret + "\",\"audience\":\"" + config.Auth.API.Audience + "\",\"grant_type\":\"client_credentials\"}")

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}
	if res == nil {
		log.Fatalf("Response is nil")
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	// --- Pretty Print Logic ---
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, body, "", "  ")
	if error != nil {
		log.Fatalf("JSON parse error: %v", error)
		// If it's not valid JSON, just print the raw body
		fmt.Println(string(body))
		return
	}

	handleAuthResponse(body)

}
