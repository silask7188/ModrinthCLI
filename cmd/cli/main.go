package main

import (
	"context"
	"fmt"
	"log"
	"silas/modrinth/modrinth"
)

func main() {
	client, err := modrinth.New("https://api.modrinth.com/v2/")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	sr, err := client.Search(ctx, modrinth.SearchParams{
		Query: "sodium",
		Limit: 22,
	})
	if err != nil {
		log.Fatal(err)
	}

	for i, p := range sr.Hits {
		fmt.Printf("%2d. %-30s %8d downloads\n", i+1, p.Title, p.Downloads)
	}
}

