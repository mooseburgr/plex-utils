package api

import (
	"github.com/jrudio/go-plex-client"
	"log"
	"net/http"
)

func initPlexCxn() (*plex.Plex, error) {
	return plex.New("https://plex.tv", "TODO-token")
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
