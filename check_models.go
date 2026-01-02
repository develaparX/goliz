package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

func main() {
	godotenv.Load()

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}

	// List models
	fmt.Println("Listing available models...")
	page, err := client.Models.List(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	for {
		for _, m := range page.Items {
			fmt.Printf("Model: %s\n", m.Name)
			fmt.Printf("  SupportedActions: %v\n", m.SupportedActions)
			fmt.Println("---")
		}

		if page.NextPageToken == "" {
			break
		}

		// Get next page
		page, err = page.Next(ctx)
		if err != nil {
			log.Println("Error fetching next page:", err)
			break
		}
	}
}
