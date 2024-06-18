package config_test

import (
	"path/filepath"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"gotest.tools/v3/assert"
)

func TestLoadServerConfig(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic")
		}
	}()

	serverConfFileName := filepath.Join(testDataDirectory, "server_test_conf.json")

	sFileConf := config.LoadServerConfigFromFile(serverConfFileName)

	aConf := &config.ServerConfiguration{
		URL:     "localhost:8082",
		Restore: false,
	}

	config.UpdateServerDefaultValues(sFileConf, aConf, true)

	assert.Equal(t, aConf.URL, "localhost:8082")
	assert.Equal(t, aConf.Restore, false)
}
