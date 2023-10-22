package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jrudio/go-plex-client"
	"github.com/pkg/errors"
	"log/slog"
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
	errKey       = "err"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

var ipInfoMap = make(map[string]IpResponse)

func SendInvite(w http.ResponseWriter, r *http.Request) {
	// CORS is COOL
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// handle OPTIONS request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var body RequestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logRequest(body, r)

	err = doThePlexStuff(body.Email)

	if err != nil {
		logger.ErrorContext(context.Background(), "err from plex",
			errKey, err)
		if strings.HasPrefix(err.Error(), strconv.Itoa(http.StatusUnprocessableEntity)) {
			// 422 == invite is already pending or user exists
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		} else {
			http.Error(w, "lol idk something blew up: "+err.Error(), http.StatusBadRequest)
		}
	} else {
		postToSlack(body.Email, GetIpAddress(r))
	}
}

func doThePlexStuff(email string) error {
	plexClient, err := initPlexClient()
	if err != nil {
		return errors.Wrap(err, "failed to connect to plex")
	}

	err = cancelAnyPendingInvites(plexClient, email)
	if err != nil {
		return err
	}

	machineId, err := plexClient.GetMachineID()
	if err != nil {
		return err
	}

	err = plexClient.InviteFriend(plex.InviteFriendParams{
		UsernameOrEmail: email,
		MachineID:       machineId,
	})
	if err != nil {
		return err
	}

	err = excludePrivateLabel(plexClient, email)
	if err != nil {
		return err
	}

	// go ensureAllHaveDownloadAccess(plexClient)

	return nil
}

func postToSlack(email, ip string) {
	ipInfo, err := GetIpInfo(ip)

	msg := fmt.Sprintf("invited to Plex: `%s` in %s, %s (`%s`)", email,
		ipInfo.City, ipInfo.RegionCode, ipInfo.Ip)

	resp, err := http.Post(os.Getenv("SLACK_WEBHOOK_URL"), "application/json",
		strings.NewReader(fmt.Sprintf(`{"text":"%s"}`, msg)))

	logger.InfoContext(context.Background(), "slack response",
		"resp", fmt.Sprintf("%+v", resp),
		errKey, err)
}

func excludePrivateLabel(cxn *plex.Plex, email string) error {
	friends, err := cxn.GetFriends()
	if err != nil {
		return errors.Wrap(err, "failed to get current friends")
	}
	for _, friend := range friends {
		if strings.EqualFold(friend.Email, email) {
			success, err := cxn.UpdateFriendAccess(fmt.Sprint(friend.ID), plex.UpdateFriendParams{
				FilterTelevision: "label!=private",
				FilterMusic:      "label!=private",
				FilterPhotos:     "label!=private",
				FilterMovies:     "label!=private",
			})
			logger.InfoContext(context.Background(), "updated friend access",
				"friend", friend,
				"success", success,
				errKey, err)
		}
	}
	return nil
}

func ensureAllHaveDownloadAccess(cxn *plex.Plex) {
	friends, err := cxn.GetFriends()
	if err != nil {
		logger.ErrorContext(context.Background(), "failed to get current friends",
			errKey, err)
	}
	for _, friend := range friends {
		success, err := cxn.UpdateFriendAccess(fmt.Sprint(friend.ID), plex.UpdateFriendParams{
			AllowSync:     "1",
			AllowChannels: "1",
		})
		if !success || err != nil {
			logger.ErrorContext(context.Background(), "failed to allow downloads for",
				"friend", friend,
				errKey, err)
		}
	}
}

func logRequest(body RequestBody, r *http.Request) {
	info, err := GetIpInfo(GetIpAddress(r))
	logger.InfoContext(context.Background(), "parsed request",
		"body", body,
		UserAgent, r.Header[UserAgent],
		"ipInfo", info,
		errKey, err)
}

func GetIpAddress(r *http.Request) string {
	if len(r.Header[XFF]) > 0 {
		return r.Header[XFF][0]
	}
	return ""
}

func cancelAnyPendingInvites(plexCxn *plex.Plex, email string) error {
	invites, err := plexCxn.GetInvitedFriends()
	if err != nil {
		return errors.Wrap(err, "failed to get invites")
	}
	for _, invite := range invites {
		if strings.EqualFold(invite.ID, email) || strings.EqualFold(invite.Email, email) {
			success, err := plexCxn.RemoveInvitedFriend(invite.ID, invite.IsFriend, invite.IsServer, invite.IsHome)
			if !success || err != nil {
				logger.ErrorContext(context.Background(), "failed to cancel pending invite",
					"invite", invite,
					errKey, err)
			} else {
				logger.InfoContext(context.Background(), "successfully canceled invite",
					"invite", invite)
			}
		}
	}
	return nil
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
		logger.ErrorContext(context.Background(), "error from IP API",
			errKey, err)
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		logger.ErrorContext(context.Background(), "failed to decode JSON",
			errKey, err)
	}
	ipInfoMap[ip] = response
	return response, nil
}

func initPlexClient() (*plex.Plex, error) {
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
