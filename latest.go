package registry

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

const LatestReleaseURL = "https://api.github.com/repos/OpenRVS-devs/OpenRVS/releases/latest"

var DefaultVersion = []byte("unknown")

type GithubResponse struct {
	TagName string `json:"tag_name"`
}

func GetLatestReleaseVersion() []byte {
	resp, err := http.Get(LatestReleaseURL)
	if err != nil {
		log.Println("error getting version from github:", err)
		return DefaultVersion
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("error reading response body:", err)
		return DefaultVersion
	}

	var r GithubResponse
	if err := json.Unmarshal(bytes, &r); err != nil {
		log.Println("error unmarshaling json:", err)
		return DefaultVersion
	}

	return []byte(r.TagName)
}
