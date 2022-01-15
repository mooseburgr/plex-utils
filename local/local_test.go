package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocal(t *testing.T) {
	t.Skip()
	main()
}

func TestDisableVpn(t *testing.T) {
	t.Skip()
	disableVpn(&gin.Context{})
	assert.True(t, true)
}

func TestEnableVpn(t *testing.T) {
	t.Skip()
	enableVpn(&gin.Context{})
	assert.True(t, true)
}
