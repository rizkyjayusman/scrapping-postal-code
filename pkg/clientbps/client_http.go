package clientbps

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type HTTP struct {
	RestClient *http.Client
	BaseUrl    *url.URL
}

func NewHTTP() (Client, error) {
	baseUrl := os.Getenv("BPS_BASE_URL")
	if baseUrl == "" {
		log.Fatal("value of bps base url is empty")
	}

	schemeUrl := os.Getenv("BPS_SCHEME_URL")
	if schemeUrl == "" {
		log.Fatal("value of bps scheme url is empty")
	}

	client := http.DefaultClient
	client.Timeout = 2 * time.Second

	e := &HTTP{
		BaseUrl: &url.URL{
			Scheme: schemeUrl,
			Host:   baseUrl,
		},
		RestClient: client,
	}
	return e, nil
}

func (e *HTTP) GetRegion(parent string, level int) (out []ResponseBodyGetRegion, err error) {
	params := url.Values{
		"parent": {parent},
	}

	if level == 1 {
		params.Add("level", "provinsi")
	} else if level == 2 {
		params.Add("level", "kabupaten")
	} else if level == 3 {
		params.Add("level", "kecamatan")
	} else {
		params.Add("level", "desa")
	}

	// Create a new GET request
	e.BaseUrl.Path = "/rest-bridging-pos/getwilayah"
	req, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", e.BaseUrl.String(), params.Encode()), nil)
	if err != nil {
		return
	}

	// Send the request
	resp, err := e.RestClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Print the response
	fmt.Println(string(body))
	fmt.Println("=================================================")

	err = json.Unmarshal([]byte(body), &out)
	if err != nil {
		return
	}

	return
}
