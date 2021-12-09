package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigs_ParseConfigFile(t *testing.T) {
	conf := &Configs{
		DebugMode: true,
		DB: &DBConfigs{
			Host: "127.0.0.1",
		},
	}

	conf.ParseConfigFile()

	assert.Equal(t, true, conf.DebugMode)
	assert.Equal(t, "127.0.0.1", conf.DB.Host)
}
