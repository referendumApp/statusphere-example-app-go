// cmd/test-atproto/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/referendumApp/statusphere-example-app-go/internal/atproto"
	"github.com/referendumApp/statusphere-example-app-go/internal/config"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get AT Protocol configuration
	cfg := config.GetATProtoConfig()
	fmt.Printf("Using PDS host: %s\n", cfg.PdsHost)

	// Create AT Protocol client
	client, err := atproto.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating AT Protocol client: %v", err)
	}

	fmt.Println("AT Protocol client created successfully")

	// Check if username and password were provided
	if len(os.Args) >= 3 {
		username := os.Args[1]
		password := os.Args[2]

		// Attempt to login
		ctx := context.Background()
		if err := client.Login(ctx, username, password); err != nil {
			log.Fatalf("Error logging in: %v", err)
		}

		fmt.Println("Successfully logged in!")

		// If a handle to look up was provided
		if len(os.Args) >= 4 {
			handle := os.Args[3]
			profile, err := client.GetProfile(ctx, handle)
			if err != nil {
				log.Fatalf("Error fetching profile for %s: %v", handle, err)
			}

			fmt.Printf("Profile for %s:\n", handle)
			fmt.Printf("  DID: %s\n", profile.Did)
			fmt.Printf("  Display Name: %s\n", profile.DisplayName)
			fmt.Printf("  Description: %s\n", profile.Description)
			fmt.Printf("  Followers: %d\n", profile.FollowersCount)
			fmt.Printf("  Following: %d\n", profile.FollowsCount)
		}
	}
}