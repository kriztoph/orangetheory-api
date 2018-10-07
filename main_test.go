package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"time"
)

func TestAuth(t *testing.T) {
	response, phpSessionID, err := auth("brentwood-norcal", "kriz@krizinc.com", "Hupwar44")
	assert.NoError(t, err)
	assert.Equal(t, "Kriztopher", response.Name)
	assert.Equal(t, "success", response.Status)
	assert.NotEqual(t, "", phpSessionID)

	scheduleResponse, err := request(time.Now(), "brentwood-norcal", phpSessionID)
	assert.NoError(t, err)
	assert.Equal(t, "", scheduleResponse)
	//bookClass("15019", "brentwood-norcal", phpSessionID)
}
