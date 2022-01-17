package api

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendInvite(t *testing.T) {
	t.Skip("local ad-hoc only")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/",
		strings.NewReader(`{"email": "joh08227@umn.edu"}`))

	SendInvite(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}
