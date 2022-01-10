package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocal(t *testing.T) {
	main()

}

func TestDisableVpn(t *testing.T) {
	disableVpn(&gin.Context{})
	assert.True(t, true)
}

func TestEnableVpn(t *testing.T) {
	enableVpn(&gin.Context{})
	assert.True(t, true)
}
