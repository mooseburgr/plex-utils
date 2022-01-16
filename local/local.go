package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/jrudio/go-plex-client"
	"github.com/mooseburgr/plex-utils/api"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	UserAgent = "User-Agent"
)

var (
	plexCxn   *plex.Plex
	zapLogger *zap.Logger
	logger    *zap.SugaredLogger
)

func init() {
	zapLogger, _ = zap.NewProduction()
	logger = zapLogger.Sugar()
}

func main() {
	defer zapLogger.Sync()

	plexCxn, err := plex.New("http://127.0.0.1:32400", api.GetPlexToken())
	if err != nil {
		panic(err)
	}

	setupNotifications(plexCxn)

	setupScheduledTasks()

	// lol all of this is meaningless after VPN split tunneling  is working
	setupRouter().Run("localhost:42069")
}

func setupScheduledTasks() {
	scheduler := gocron.NewScheduler(time.Local)
	// reconnect after router reboots
	scheduler.Every(1).Days().At("04:30").Do(enableVpn)
	// disable
	scheduler.Every(1).Days().At("07:30").Do(disableVpn)
}

func setupNotifications(cxn *plex.Plex) {
	logger.Debug(plexCxn.GetMachineID())
	ctrlC := make(chan os.Signal, 1)
	onError := func(err error) {
		logger.Errorf("error from event: %v", err)
	}

	events := plex.NewNotificationEvents()
	events.OnPlaying(func(n plex.NotificationContainer) {
		disableVpn(nil)
		mediaID := n.PlaySessionStateNotification[0].RatingKey
		sessionID := n.PlaySessionStateNotification[0].SessionKey

		sessions, err := cxn.GetSessions()
		if err != nil {
			logger.Errorf("failed to fetch sessions on plex server: %v", err)
			return
		}

		var userID, username, title string
		for _, meta := range sessions.MediaContainer.Metadata {
			if sessionID != meta.SessionKey {
				continue
			}
			userID = meta.User.ID
			username = meta.User.Title
			break
		}

		metadata, err := cxn.GetMetadata(mediaID)
		if err != nil {
			logger.Warnf("failed to get metadata for key %s: %v\n", mediaID, err)
		} else {
			title = metadata.MediaContainer.Metadata[0].Title
		}
		logger.Infof("user %s (id: %s) has started playing %s (id: %s)", username, userID,
			title, mediaID)
	})
	cxn.SubscribeToNotifications(events, ctrlC, onError)
}

func setupWebhooks() *plex.WebhookEvents {
	webhooks := plex.NewWebhook()
	webhooks.OnPlay(func(w plex.Webhook) {
		logger.Debugw("account %v is playing %v", w.Account.Title, w.Metadata.Title)
	})
	webhooks.OnResume(func(w plex.Webhook) {
		logger.Debugw("account %v resumed %v", w.Account.Title, w.Metadata.Title)
	})
	return webhooks
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Use(logRequest)
	router.LoadHTMLGlob("templates/*.tmpl")

	router.GET("/favicon.ico", getFavicon)
	router.GET("/admin", func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin.tmpl", gin.H{
			"vpnEnabled": isVpnEnabled(),
		})
	})
	router.POST("/vpn/disable.do", disableVpn)
	router.POST("/vpn/enable.do", enableVpn)
	return router
}

func disableVpn(c *gin.Context) {
	// pause all torrents (https://github.com/fedarovich/qbittorrent-cli/wiki/command-reference)
	out, err := exec.Command("qbt", "torrent", "pause", "ALL").Output()
	logOutput("pausing torrents", out, err)

	// lol idk, wait a few secs for cxns to close??
	time.Sleep(2 * time.Second)

	// disconnect VPN
	out, err = exec.Command("nordvpn", "-d").Output()
	logOutput("disconnecting VPN", out, err)
}

func enableVpn(c *gin.Context) {
	// connect VPN
	out, err := exec.Command("nordvpn", "-c").Output()
	logOutput("connecting VPN", out, err)

	// kill torrent client
	out, err = exec.Command("taskkill", "/F", "/IM", "qbittorrent.exe").Output()
	logOutput("killing client", out, err)

	// give VPN some time to connect
	time.Sleep(2 * time.Second)

	// restart torrent client
	restartCmd := exec.Command("qbittorrent")
	restartCmd.Dir = "E:/Program Files/qBittorrent/"
	err = restartCmd.Start()
	logOutput("starting client", nil, err)

	// resume all torrents
	out, err = exec.Command("qbt", "torrent", "resume", "ALL").Output()
	logOutput("resuming torrents", out, err)
}

func logRequest(c *gin.Context) {
	logger.Infow("Handling request",
		"method", c.Request.Method,
		"path", c.FullPath(),
		"ip", c.ClientIP(),
		UserAgent, c.GetHeader(UserAgent))
}

func logOutput(msg string, out []byte, err error) {
	logger.Infow(msg,
		"out", string(out),
		"err", err)
}

func isVpnEnabled() bool {
	// TODO fix this crap
	resp, err := http.Get("https://ipapi.co/org/")
	if err != nil {
		logger.Warnf("error from IP API: %v", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("failed to read IP API response: %v", err)
	}
	return !strings.Contains(string(b), "CENTURYLINK")
}

func getFavicon(c *gin.Context) {
	resp, _ := http.Get("https://s.gravatar.com/avatar/0dcd9557e311bb567da7dad218069b76")
	reader := resp.Body
	defer reader.Close()
	contentLength := resp.ContentLength
	contentType := resp.Header.Get("Content-Type")
	extraHeaders := map[string]string{
		"Content-Disposition": `attachment; filename="favicon.png"`,
	}
	c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
}
