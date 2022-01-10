package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jrudio/go-plex-client"
	"go.uber.org/zap"
	"net/http"
	"os/exec"
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

	plexCxn, err := plex.New("http://127.0.01:32400", "TODO-get-me")
	if err != nil {
		panic(err)
	}

	logger.Info(plexCxn.GetMachineID())
	//webhooks := plex.NewWebhook()
	//println(webhooks)
	//events := plex.NewNotificationEvents()
	//println(events)

	// lol all of this is meaningless after remote access with VPN is fixed
	setupRouter().Run("localhost:42069")
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Use(logRequest)
	router.LoadHTMLGlob("templates/*.tmpl")

	router.GET("/cookie", func(c *gin.Context) {
		c.SetCookie("gin_cookie", c.ClientIP(), 3600, "/", "localhost", false, true)
	})
	router.GET("/admin", func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin.tmpl", gin.H{
			"key": "value",
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

	// idk, wait a few secs for cxns to close??
	time.Sleep(3 * time.Second)

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
	time.Sleep(3 * time.Second)

	// restart torrent client
	restartCmd := exec.Command("qbittorrent")
	restartCmd.Dir = "E:/Program Files/qBittorrent/"
	err = restartCmd.Start()
	logOutput("starting client", out, err)

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
