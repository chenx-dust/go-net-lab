package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// JSONConfig represents the JSON structure that matches Config
type JSONConfig struct {
	Mode         string        `json:"mode"`
	ListenAddr   string        `json:"listen_addr"`
	RemoteAddr   string        `json:"remote_addr,omitempty"`
	RelayServers []RelayServer `json:"relay_servers,omitempty"`
	RelayType    *RelayType    `json:"relay_type,omitempty"`
}

// LoadFromFile reads and parses a JSON configuration file
func LoadFromFile(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var jsonConfig JSONConfig
	if err := json.Unmarshal(data, &jsonConfig); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return convertJSONConfig(jsonConfig)
}

// convertJSONConfig converts JSONConfig to Config
func convertJSONConfig(jc JSONConfig) (*Config, error) {
	// Convert mode string to AppMode
	var mode AppMode
	switch jc.Mode {
	case "client":
		mode = ClientMode
	case "relay":
		mode = RelayMode
	case "server":
		mode = ServerMode
	default:
		return nil, fmt.Errorf("invalid mode: %s", jc.Mode)
	}

	config := &Config{
		Mode:         mode,
		ListenAddr:   jc.ListenAddr,
		RemoteAddr:   jc.RemoteAddr,
		RelayServers: jc.RelayServers,
	}

	if jc.RelayType != nil {
		config.RelayType = *jc.RelayType
	}

	return config, nil
}
