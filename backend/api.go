package api

import (
	"encoding/json"
	"fmt"
	"github.com/jrudio/go-plex-client"
	"log"
	"net/http"
	"os"
	"strings"
)

func initPlexCxn() (*plex.Plex, error) {
	return plex.New("https://plex.tv", GetPlexToken())
}

func SendInvite(w http.ResponseWriter, r *http.Request) {
	// handle OPTIONS request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var body RequestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("parsed body: %v", body)

	plexCxn, err := initPlexCxn()
	if err != nil {
		http.Error(w, "ruh roh, can't connect to plex: "+err.Error(), http.StatusInternalServerError)
		return
	}

	cancelAnyPendingInvites(plexCxn, body.Email)

	err = plexCxn.InviteFriend(plex.InviteFriendParams{
		UsernameOrEmail: body.Email,
		MachineID:       "d92d03d0c5f98de89a3b7699d744949bd9e78424",
	})

	if err != nil {
		log.Printf("err from plex: %v", err)
		if strings.HasPrefix(err.Error(), fmt.Sprint(http.StatusUnprocessableEntity)) {
			// 422 = invite is already pending or user exists
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		} else {
			http.Error(w, "lol idk something went wrong: "+err.Error(), http.StatusBadRequest)
		}
	}
}

func cancelAnyPendingInvites(plexCxn *plex.Plex, email string) {
	invites, _ := plexCxn.GetInvitedFriends()
	for _, invite := range invites {
		if invite.ID == email || invite.Email == email {
			plexCxn.RemoveInvitedFriend(invite.ID, invite.IsFriend, invite.IsServer, invite.IsHome)
		}
	}
}

type RequestBody struct {
	Email string `json:"email"`
}

func GetPlexToken() string {
	if token := os.Getenv("PLEX_TOKEN"); token != "" {
		return token
	} else {
		bytes, _ := os.ReadFile("plex-token")
		return string(bytes)
	}
}
