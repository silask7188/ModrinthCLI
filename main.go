package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"log"
	"silas/modrinth/modrinth"

)



// @brief   Main!
// @details This is the main function that starts the program.
// @return  void
func main() {
	resp, err := http.Get("https://api.modrinth.com/v2/search")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("bad status: %s", resp.Status)
	}
	
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

    var sr modrinth.SearchResponse
	if err := json.Unmarshal(raw, &sr); err != nil {
		log.Fatal(err)
	}
	
	for i, p := range sr.Hits {
		fmt.Printf("%2d. %s (%d downloads)\n", i+1, p.Title, p.Downloads)
	}
}
