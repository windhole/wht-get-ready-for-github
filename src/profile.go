package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed profiles/*.json
var profilesFS embed.FS

type profileConfig struct {
	License    bool   `json:"license"`
	Visibility string `json:"visibility"`
}

func loadProfile(name string) (profileConfig, error) {
	raw, err := profilesFS.ReadFile("profiles/" + name + ".json")
	if err != nil {
		return profileConfig{}, fmt.Errorf("プロファイル %q が見つかりません", name)
	}
	var cfg profileConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return profileConfig{}, fmt.Errorf("プロファイル %q の JSON が不正です: %w", name, err)
	}
	switch cfg.Visibility {
	case "private", "public":
	default:
		return profileConfig{}, fmt.Errorf("プロファイル %q の visibility が不正です: %s", name, cfg.Visibility)
	}
	return cfg, nil
}

func listProfileNames() []string {
	entries, err := fs.ReadDir(profilesFS, "profiles")
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".json") {
			names = append(names, strings.TrimSuffix(name, ".json"))
		}
	}
	sort.Strings(names)
	return names
}
