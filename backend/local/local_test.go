package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mooseburgr/plex-utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocal(t *testing.T) {
	t.Skip()
	backend.main()
}

func TestDisableVpn(t *testing.T) {
	t.Skip()
	backend.disableVpn(&gin.Context{})
	assert.True(t, true)
}

func TestEnableVpn(t *testing.T) {
	t.Skip()
	backend.enableVpn(&gin.Context{})
	assert.True(t, true)
}
