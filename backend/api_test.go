package api

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"
)

func Test_SendInvite(t *testing.T) {
	if !isLocal() {
		t.Skip()
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/",
		strings.NewReader(`{"email": "joh08227@umn.edu"}`))

	SendInvite(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_GetIpInfo(t *testing.T) {
	ip := "76.223.122.69"
	info, err := GetIpInfo(ip)

	assert.NoError(t, err)
	assert.Equal(t, ip, info.Ip)

	info2, err := GetIpInfo(ip)

	assert.NoError(t, err)
	assert.Equal(t, info, info2)
}

func Test_postToSlack(t *testing.T) {
	if !isLocal() {
		t.Skip()
	}
	os.Setenv("SLACK_WEBHOOK_URL", "https://hooks.slack.com/services/SUPER/SECRET/URL")
	postToSlack("test@invite.com", "76.223.122.69")
}

func isLocal() bool {
	host, _ := os.Hostname()
	return slices.Contains([]string{"Kyles-MBP", "Kyles-MacBook-Pro.local", "Alakazam11"}, host)
}
