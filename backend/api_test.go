package api

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func Test_SendInvite(t *testing.T) {
	t.Skip("local ad-hoc only")
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
	t.Skip("local ad-hoc only")
	os.Setenv("SLACK_WEBHOOK_URL", "https://hooks.slack.com/services/")
	postToSlack("test@invite.com")
}
