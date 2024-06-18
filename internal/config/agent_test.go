package config_test

import (
	"path/filepath"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"gotest.tools/v3/assert"
)

const testDataDirectory = "../../testdata"

func TestLoadAgentConfig(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic")
		}
	}()

	agentConfFileName := filepath.Join(testDataDirectory, "agent_test_conf.json")

	aFileConf := config.LoadAgentConfigFromFile(agentConfFileName)

	aConf := &config.AgentConfiguration{
		ServerAddr:   "localhost:8082",
		PollInterval: config.AgentDefautlPollInterval,
	}

	config.UpdateAgentDefaultValues(aFileConf, aConf)

	assert.Equal(t, "localhost:8082", aConf.ServerAddr)
	assert.Equal(t, 1, aConf.PollInterval)
}
