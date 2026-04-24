package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"docker-pilot/internal/system"
	"docker-pilot/internal/ui"
)

// Default configuration - customize for your enterprise here
const (
	DefaultRegistry   = "registry.example.com"
	DefaultHTTPProxy  = "http://proxy.example.com:8080"
	DefaultHTTPSProxy = "http://proxy.example.com:8080"
	DefaultNoProxy    = "localhost,127.0.0.1,.example.com"
	DefaultBIP        = "172.31.0.1/16"
	SkipKeyword       = "skip"
)

// RegistryConfig holds registry configuration
type RegistryConfig struct {
	Configured bool
	Registry   string
}

// ProxyConfig holds HTTP proxy configuration
type ProxyConfig struct {
	Configured bool
	HTTPProxy  string
	HTTPSProxy string
	NoProxy    string
}

// CIDRConfig holds container network CIDR configuration
type CIDRConfig struct {
	Configured bool
	BIP        string
}

// DaemonConfig corresponds to /etc/docker/daemon.json
type DaemonConfig struct {
	InsecureRegistries []string `json:"insecure-registries,omitempty"`
	RegistryMirrors    []string `json:"registry-mirrors,omitempty"`
	BIP                string   `json:"bip,omitempty"`
}

// AskRegistryConfig prompts for registry configuration
func AskRegistryConfig() (RegistryConfig, error) {
	help := "Without registry, internal images cannot be pulled. Enter " + SkipKeyword + " to skip temporarily"
	answer, err := ui.AskInput("Internal registry address", DefaultRegistry, help)
	if err != nil {
		return RegistryConfig{}, err
	}

	if strings.ToLower(answer) == SkipKeyword {
		return RegistryConfig{Configured: false}, nil
	}

	return RegistryConfig{
		Configured: true,
		Registry:   strings.TrimSpace(answer),
	}, nil
}

// AskProxyConfig prompts for proxy configuration
func AskProxyConfig() (ProxyConfig, error) {
	help := "Proxy required for external network access. Enter " + SkipKeyword + " to skip temporarily"

	httpProxy, err := ui.AskInput("HTTP proxy address", DefaultHTTPProxy, help)
	if err != nil {
		return ProxyConfig{}, err
	}

	if strings.ToLower(httpProxy) == SkipKeyword {
		return ProxyConfig{Configured: false}, nil
	}

	httpsProxy, err := ui.AskInput("HTTPS proxy address", DefaultHTTPSProxy, "")
	if err != nil {
		return ProxyConfig{}, err
	}

	noProxy, err := ui.AskInput("NO_PROXY - domains to bypass proxy", DefaultNoProxy, "")
	if err != nil {
		return ProxyConfig{}, err
	}

	return ProxyConfig{
		Configured: true,
		HTTPProxy:  strings.TrimSpace(httpProxy),
		HTTPSProxy: strings.TrimSpace(httpsProxy),
		NoProxy:    strings.TrimSpace(noProxy),
	}, nil
}

// AskCIDRConfig prompts for container network CIDR configuration
func AskCIDRConfig() (CIDRConfig, error) {
	options := []string{
		"172.31.0.0/16 (Recommended, avoids conflicts with most networks)",
		"10.200.0.0/16",
		"192.168.100.0/24",
		"Custom input",
	}

	choice, err := ui.AskSelect("Select container network CIDR (avoid conflicts with internal network)", options, 0)
	if err != nil {
		return CIDRConfig{}, err
	}

	var bip string
	if choice == 3 {
		// Custom input
		bip, err = ui.AskInput("Enter custom CIDR (e.g., 172.31.0.1/16)", DefaultBIP, "")
		if err != nil {
			return CIDRConfig{}, err
		}
	} else {
		// Extract CIDR from selected option
		bip = strings.Fields(options[choice])[0]
		// Convert network address to gateway address, e.g., 172.31.0.0/16 -> 172.31.0.1/16
		bip = strings.Replace(bip, ".0/", ".1/", 1)
	}

	// Verify if CIDR is already occupied
	if system.IsCIDROccupied(bip) {
		ui.PrintWarning("This CIDR may already be in use, consider selecting another")
		if !ui.AskConfirm("Are you sure you want to use this CIDR?", false) {
			// Ask again
			return AskCIDRConfig()
		}
	}

	return CIDRConfig{
		Configured: true,
		BIP:        bip,
	}, nil
}

// WriteRegistryConfig writes registry configuration to daemon.json
func WriteRegistryConfig(config RegistryConfig) error {
	if err := system.BackupFile("/etc/docker/daemon.json"); err != nil {
		ui.PrintWarning("Failed to backup daemon.json")
	}

	// Read existing configuration
	daemonConfig, err := readDaemonConfig()
	if err != nil {
		daemonConfig = &DaemonConfig{}
	}

	// Add registry configuration
	daemonConfig.InsecureRegistries = appendUnique(daemonConfig.InsecureRegistries, config.Registry)
	daemonConfig.RegistryMirrors = appendUnique(daemonConfig.RegistryMirrors, "https://"+config.Registry)

	return writeDaemonConfig(daemonConfig)
}

// WriteProxyConfig writes proxy configuration to systemd drop-in file
func WriteProxyConfig(config ProxyConfig) error {
	dropinDir := "/etc/systemd/system/docker.service.d"
	if err := os.MkdirAll(dropinDir, 0755); err != nil {
		return err
	}

	content := `[Service]
Environment="HTTP_PROXY=` + config.HTTPProxy + `"
Environment="HTTPS_PROXY=` + config.HTTPSProxy + `"
Environment="NO_PROXY=` + config.NoProxy + `"
`

	return os.WriteFile(filepath.Join(dropinDir, "http-proxy.conf"), []byte(content), 0644)
}

// WriteCIDRConfig writes CIDR configuration to daemon.json
func WriteCIDRConfig(config CIDRConfig) error {
	if err := system.BackupFile("/etc/docker/daemon.json"); err != nil {
		ui.PrintWarning("Failed to backup daemon.json")
	}

	daemonConfig, err := readDaemonConfig()
	if err != nil {
		daemonConfig = &DaemonConfig{}
	}

	daemonConfig.BIP = config.BIP

	return writeDaemonConfig(daemonConfig)
}

// readDaemonConfig reads daemon.json
func readDaemonConfig() (*DaemonConfig, error) {
	data, err := os.ReadFile("/etc/docker/daemon.json")
	if err != nil {
		return nil, err
	}

	var config DaemonConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// writeDaemonConfig writes daemon.json
func writeDaemonConfig(config *DaemonConfig) error {
	if err := os.MkdirAll("/etc/docker", 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("/etc/docker/daemon.json", data, 0644)
}

// appendUnique appends an item if not already present
func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}
