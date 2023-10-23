package initializers

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewLogger(t *testing.T) {
	// Test with debug log level environment variable
	err := os.Setenv(LOG_LEVEL_ENV, "debug")
	if err != nil {
		t.Errorf("Unable to set environment variable.")
	}
	logger, err := NewLogger(true, true)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
}
