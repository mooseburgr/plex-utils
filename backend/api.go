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

const (
	UserAgent  = "User-Agent"
	IpStackKey = "dba9b8dc10f06971ee169e857c374d07" // free key, wgaf
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
	info, _ := GetIpInfo(r.RemoteAddr)
	log.Printf("parsed body: %v \nfrom IP: %+v \nuser-agent: %v", body, info, r.Header[UserAgent])

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
			success, err := plexCxn.RemoveInvitedFriend(invite.ID, invite.IsFriend, invite.IsServer, invite.IsHome)
			if !success || err != nil {
				log.Printf("failed to cancel pending invite %v, err: %v", invite, err)
			} else {
				log.Printf("successfully canceled invite %v", invite)
			}
		}
	}
}

type RequestBody struct {
	Email string `json:"email"`
}

type IpResponse struct {
	Ip            string  `json:"ip"`
	Type          string  `json:"type"`
	ContinentCode string  `json:"continent_code"`
	ContinentName string  `json:"continent_name"`
	CountryCode   string  `json:"country_code"`
	CountryName   string  `json:"country_name"`
	RegionCode    string  `json:"region_code"`
	RegionName    string  `json:"region_name"`
	City          string  `json:"city"`
	Zip           string  `json:"zip"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Location      struct {
		GeonameId int    `json:"geoname_id"`
		Capital   string `json:"capital"`
		Languages []struct {
			Code   string `json:"code"`
			Name   string `json:"name"`
			Native string `json:"native"`
		} `json:"languages"`
		CountryFlag             string `json:"country_flag"`
		CountryFlagEmoji        string `json:"country_flag_emoji"`
		CountryFlagEmojiUnicode string `json:"country_flag_emoji_unicode"`
		CallingCode             string `json:"calling_code"`
		IsEu                    bool   `json:"is_eu"`
	} `json:"location"`
}

func GetIpInfo(ip string) (IpResponse, error) {
	var response IpResponse
	resp, err := http.Get(fmt.Sprintf("http://api.ipstack.com/%v?access_key=%v", ip, IpStackKey))
	if err != nil {
		log.Printf("error from IP API: %v", err)
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Printf("error decoding: %v", err)
	}
	return response, nil
}

func GetPlexToken() string {
	if token := os.Getenv("PLEX_TOKEN"); token != "" {
		return token
	} else {
		bytes, _ := os.ReadFile("plex-token")
		return string(bytes)
	}
}
