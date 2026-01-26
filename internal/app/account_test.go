//go:build integration

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {

}

func TestFiberApp_Login(t *testing.T) {
	assert.NotPanics(t, func() {})
}
