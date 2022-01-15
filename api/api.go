package api

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"context"
	"fmt"
	"github.com/jrudio/go-plex-client"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"log"
	"net/http"
	"os"
)

func initPlexCxn() (*plex.Plex, error) {
	return plex.New("https://plex.tv", GetPlexToken())
}

func SendInvite(w http.ResponseWriter, r *http.Request) {

	email := r.URL.Query().Get("email")
	plexCxn, err := initPlexCxn()
	if err != nil {
		log.Print(err)
		http.Error(w, "ruh roh, can't connect to plex: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = plexCxn.InviteFriend(plex.InviteFriendParams{
		UsernameOrEmail: email,
		MachineID:       "d92d03d0c5f98de89a3b7699d744949bd9e78424",
	})

	if err != nil {
		log.Print(err)
		http.Error(w, "lol idk something went wrong: "+err.Error(), http.StatusBadRequest)
	}
}

type requestBody struct {
	email string `json:"email"`
}

func GetPlexToken() string {
	gcpProjectId := os.Getenv("GCP_PROJECT")
	if gcpProjectId == "" {
		bytes, _ := os.ReadFile("plex-token")
		return string(bytes)
	} else {
		// Create the client.
		ctx := context.Background()
		client, err := secretmanager.NewClient(ctx)
		if err != nil {
			log.Fatalf("failed to setup client: %v", err)
		}
		defer client.Close()

		// Build the request.
		accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
			Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", gcpProjectId, "TODO"),
		}

		// Call the API.
		result, err := client.AccessSecretVersion(ctx, accessRequest)
		if err != nil {
			log.Fatalf("failed to access secret version: %v", err)
		}
		return string(result.Payload.Data)
	}
}
