package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Config holds the scanner's runtime configuration. Loaded from YAML.
type Config struct {
	Host string `yaml:"host"`

	Device          string        `yaml:"device"`
	Duration        time.Duration `yaml:"duration"`
	AllowDuplicates bool          `yaml:"allow_duplicates"`
	AllowID152      bool          `yaml:"allow_id152"`
	AllowNHS        bool          `yaml:"allow_nhs"`
	LogBlanks       bool          `yaml:"log_blanks"`
	LogLevel        string        `yaml:"log_level"`

	Elastic struct {
		Servers      []string
		APIKeyID     string `yaml:"api_key_id"`
		APIKeySecret string `yaml:"api_id_secret"`
		Index        string `yaml:"index"`
	} `yaml:"elastic"`

	Output struct {
		JSONL struct {
			Enabled       bool   `yaml:"enabled"`
			Dir           string `yaml:"dir"`             // output directory
			Rotation      string `yaml:"rotation"`        // "daily" | "size" | "none"
			MaxSizeMB     int    `yaml:"max_size_mb"`     // when rotation: size
			CompressAfter int    `yaml:"compress_after"`  // gzip files older than N days; 0 = off
		} `yaml:"jsonl"`
	} `yaml:"output"`

	Dedup struct {
		Enabled bool          `yaml:"enabled"`
		Window  time.Duration `yaml:"window"` // drop identical packets within this window
	} `yaml:"dedup"`

	// Exclude drops broadcasters that match any of these rules BEFORE they
	// are aggregated, so your own household devices don't pollute the
	// capture. Two layers:
	//
	//   - MAC/name rules apply to every packet up front (fast path).
	//   - Identity-fingerprint rules apply AFTER manufacturer-data is
	//     parsed and catch devices that rotate their BLE MAC. Use these
	//     for Apple devices (phones, iPads, Watches, HomeKit accessories).
	//
	// Use the persistence query (./reporter/run.sh 09) to find always-
	// present MACs, and the fingerprint query (./reporter/run.sh 31) to
	// find identity tuples worth excluding.
	Exclude struct {
		Enabled         bool     `yaml:"enabled"`
		Addresses       []string `yaml:"addresses"`        // exact MAC match (case-insensitive)
		AddressPrefixes []string `yaml:"address_prefixes"` // OUI-style prefix match (e.g. "aa:bb:cc")
		LocalNames      []string `yaml:"local_names"`      // exact local-name match (case-insensitive)

		// Identity-based (survives MAC rotation).
		AirdropHashes    []string `yaml:"airdrop_hashes"`     // "appleid:phone:email:email2" tuple from airdrop subtype
		ICloudIDs        []string `yaml:"icloud_ids"`         // tethering_target.icloud_id (rotates daily)
		HomeKitDeviceIDs []string `yaml:"homekit_device_ids"` // homekit.device_id (never rotates)
		FindMyPublicKeys []string `yaml:"findmy_public_keys"` // findmy.public_key (separated mode, rotates daily)
	} `yaml:"exclude"`
}

// Excludes reports whether the given (address, local name) pair should be
// filtered out before aggregation. Returns false when exclusion is disabled.
func (c *Config) Excludes(addr, name string) bool {
	if c == nil || !c.Exclude.Enabled {
		return false
	}
	addr = strings.ToLower(addr)
	for _, x := range c.Exclude.Addresses {
		if strings.ToLower(x) == addr {
			return true
		}
	}
	for _, p := range c.Exclude.AddressPrefixes {
		if strings.HasPrefix(addr, strings.ToLower(p)) {
			return true
		}
	}
	if name != "" {
		for _, n := range c.Exclude.LocalNames {
			if strings.EqualFold(n, name) {
				return true
			}
		}
	}
	return false
}

// Fingerprints are the per-packet identifiers the exclusion layer compares
// against (after manufacturer data is parsed). Empty strings are ignored.
type Fingerprints struct {
	AirdropTuple    string // "apple_id_hash:phone_hash:email_hash:email2_hash"
	ICloudID        string // tethering_target.icloud_id
	HomeKitDeviceID string // homekit.device_id
	FindMyPublicKey string // findmy.public_key (separated mode)
}

// ExcludesByFingerprint reports whether the parsed identity should be
// dropped. Matches case-insensitively on hex content.
func (c *Config) ExcludesByFingerprint(fp Fingerprints) bool {
	if c == nil || !c.Exclude.Enabled {
		return false
	}
	if fp.AirdropTuple != "" {
		for _, x := range c.Exclude.AirdropHashes {
			if strings.EqualFold(x, fp.AirdropTuple) {
				return true
			}
		}
	}
	if fp.ICloudID != "" {
		for _, x := range c.Exclude.ICloudIDs {
			if strings.EqualFold(x, fp.ICloudID) {
				return true
			}
		}
	}
	if fp.HomeKitDeviceID != "" {
		for _, x := range c.Exclude.HomeKitDeviceIDs {
			if strings.EqualFold(x, fp.HomeKitDeviceID) {
				return true
			}
		}
	}
	if fp.FindMyPublicKey != "" {
		for _, x := range c.Exclude.FindMyPublicKeys {
			if strings.EqualFold(x, fp.FindMyPublicKey) {
				return true
			}
		}
	}
	return false
}

// LoadConfig opens path and decodes a Config from it.
func LoadConfig(path string) (*Config, error) {
	s, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
		return nil, fmt.Errorf("%q is a directory, not a normal file", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := &Config{}
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
