package main

import (
	"context"
	api "github.com/mooseburgr/plex-utils"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

func mainLocalFunc() {
	ctx := context.Background()
	if err := funcframework.RegisterHTTPFunctionContext(ctx, "/", api.SendInvite); err != nil {
		log.Fatalf("funcframework.RegisterHTTPFunctionContext: %v\n", err)
	}
	// Use PORT environment variable, or default to 8080.
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
