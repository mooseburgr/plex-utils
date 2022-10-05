package api

import (
	"encoding/json"
	"fmt"
	"github.com/jrudio/go-plex-client"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	UserAgent    = "User-Agent"
	XFF          = "X-Forwarded-For"
	AppEngUserIp = "Appengine-User-Ip"
	IpStackKey   = "dba9b8dc10f06971ee169e857c374d07" // free key, wgaf
)

var ipInfoMap = make(map[string]IpResponse)

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
	logParse(body, r)

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

	excludePrivateLabel(plexCxn, body.Email)

	// go ensureAllHaveDownloadAccess(plexCxn)

	if err != nil {
		log.Printf("err from plex: %v", err)
		if strings.HasPrefix(err.Error(), strconv.Itoa(http.StatusUnprocessableEntity)) {
			// 422 = invite is already pending or user exists
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		} else {
			http.Error(w, "lol idk something blew up: "+err.Error(), http.StatusBadRequest)
		}
	} else {
		postToSlack(body.Email, GetIpAddress(r))
	}
}

func postToSlack(email, ip string) {
	ipInfo, err := GetIpInfo(ip)

	msg := fmt.Sprintf("invited to Plex: `%s` in %s, %s (`%s`)", email,
		ipInfo.City, ipInfo.RegionCode, ipInfo.Ip)

	resp, err := http.Post(os.Getenv("SLACK_WEBHOOK_URL"), "application/json",
		strings.NewReader(fmt.Sprintf(`{"text":"%s"}`, msg)))
	log.Printf("slack response: %+v, err: %v", resp, err)
}

func excludePrivateLabel(cxn *plex.Plex, email string) {
	friends, err := cxn.GetFriends()
	if err != nil {
		log.Printf("failed to get current friends: %v", err)
	}
	for _, friend := range friends {
		if strings.EqualFold(friend.Email, email) {
			success, err := cxn.UpdateFriendAccess(fmt.Sprint(friend.ID), plex.UpdateFriendParams{
				FilterTelevision: "label!=private",
				FilterMusic:      "label!=private",
				FilterPhotos:     "label!=private",
				FilterMovies:     "label!=private",
			})
			log.Printf("updated friend %+v, success: %v, err: %v", friend, success, err)
		}
	}
}

func ensureAllHaveDownloadAccess(cxn *plex.Plex) {
	friends, err := cxn.GetFriends()
	if err != nil {
		log.Printf("failed to get current friends: %v", err)
	}
	for _, friend := range friends {
		success, err := cxn.UpdateFriendAccess(fmt.Sprint(friend.ID), plex.UpdateFriendParams{
			AllowSync:     "1",
			AllowChannels: "1",
		})
		if !success || err != nil {
			log.Printf("failed to allow downloads for: %+v", friend)
		}
	}
}

func logParse(body RequestBody, r *http.Request) {
	log.Printf("parsed body: %+v", body)
	info, err := GetIpInfo(GetIpAddress(r))
	log.Printf("from IP: from IP: %+v (err: %v)", info, err)
	log.Printf("user-agent: %v", r.Header[UserAgent])
}

func GetIpAddress(r *http.Request) string {
	return r.Header[XFF][0]
}

func cancelAnyPendingInvites(plexCxn *plex.Plex, email string) {
	invites, _ := plexCxn.GetInvitedFriends()
	for _, invite := range invites {
		if strings.EqualFold(invite.ID, email) || strings.EqualFold(invite.Email, email) {
			success, err := plexCxn.RemoveInvitedFriend(invite.ID, invite.IsFriend, invite.IsServer, invite.IsHome)
			if !success || err != nil {
				log.Printf("failed to cancel pending invite %+v, err: %v", invite, err)
			} else {
				log.Printf("successfully canceled invite %+v", invite)
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
	if info, found := ipInfoMap[ip]; found {
		return info, nil
	}

	var response IpResponse
	resp, err := http.Get(fmt.Sprintf("http://api.ipstack.com/%s?access_key=%s", ip, IpStackKey))
	if err != nil {
		log.Printf("error from IP API: %v", err)
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Printf("error decoding: %v", err)
	}
	ipInfoMap[ip] = response
	return response, nil
}

func initPlexCxn() (*plex.Plex, error) {
	return plex.New("https://plex.tv", GetPlexToken())
}

func GetPlexToken() string {
	if token := os.Getenv("PLEX_TOKEN"); token != "" {
		return token
	} else {
		bytes, _ := os.ReadFile("plex-token")
		return string(bytes)
	}
}
