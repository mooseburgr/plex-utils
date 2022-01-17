package api

import (
	"encoding/json"
	"github.com/jrudio/go-plex-client"
	"log"
	"net/http"
	"os"
)

func initPlexCxn() (*plex.Plex, error) {
	return plex.New("https://plex.tv", GetPlexToken())
}

func SendInvite(w http.ResponseWriter, r *http.Request) {
	var body requestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	plexCxn, err := initPlexCxn()
	if err != nil {
		http.Error(w, "ruh roh, can't connect to plex: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = plexCxn.InviteFriend(plex.InviteFriendParams{
		UsernameOrEmail: body.email,
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
	if token := os.Getenv("PLEX_TOKEN"); token != "" {
		return token
	} else {
		bytes, _ := os.ReadFile("plex-token")
		return string(bytes)
	}
}
