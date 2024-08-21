package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"time"

	pslices "github.com/ilknarf/piter/slices"
)

const threadLimit = 10

type Result struct {
	Response string
	Search   []*Entry `json:"Search"`
}

type Entry struct {
	Title string `json:"Title"`
	Type  string `json:"Type"`
	Year  string `json:"Year"`
}

// a list of kw to query for
var keyword = []string{
	"Friends",
	"Deadpool",
	"X-Men",
	"Arcane",
	"Citizen Kane",
	"Seinfeld",
	"I Love Lucy",
	"Severence",
	"Rings of Power",
	"Start Trek",
	"Foundation",
	"The Boys",
	"Fleabag",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
	"No more matches",
}

var ombdKey = os.Getenv("OMDB_KEY")

func sendOmdbRequest(ctx context.Context, url string) (*Result, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// JSON stuff
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &Result{}

	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func searchOmdb(ctx context.Context, kw string) ([]*Entry, error) {
	// HTTP request stuff
	url := fmt.Sprintf("https://www.omdbapi.com/?s=%s&apiKey=%s", url.QueryEscape(kw), ombdKey)

	result, err := sendOmdbRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	if result.Response != "True" {
		return nil, nil
	}

	// return []*Entry
	return result.Search, nil
}

func main() {
	// something like 1000 request/day on the free key

	// 30 second timeout for the script
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("getting results")
	res, err := pslices.PFlatMap(ctx, slices.Values(keyword), threadLimit, searchOmdb)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("printing results")

	// print results
	for _, entry := range res {
		b, err := json.Marshal(entry)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(b))
	}
}
