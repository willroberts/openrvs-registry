package registry

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// latestReleaseURL contains the latest release info for OpenRVS.
const latestReleaseURL = "https://api.github.com/repos/OpenRVS-devs/OpenRVS/releases/latest"

// githubResponse is the JSON structure we want to parse.
type githubResponse struct {
	TagName string `json:"tag_name"`
}

// GetLatestReleaseVersion retrieves the latest OpenRVS release tag from Github
// and returns it as []byte.
func GetLatestReleaseVersion() []byte {
	// Get the HTTP response.
	resp, err := http.Get(latestReleaseURL)
	if err != nil {
		log.Println("error getting version from github:", err)
		return []byte("unknown")
	}
	defer resp.Body.Close()

	// Read the response body as bytes.
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("error reading response body:", err)
		return []byte("unknown")
	}

	// Parse the JSON, storing it directly in githubResponse r.
	var r githubResponse
	if err := json.Unmarshal(bytes, &r); err != nil {
		log.Println("error unmarshaling json:", err)
		return []byte("unknown")
	}

	// Return just the latest version tag as []byte.
	return []byte(r.TagName)
}
