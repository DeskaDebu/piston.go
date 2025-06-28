package piston

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"runtime"
)

func fetchManifest() *VersionManifest {
	resp, err := http.Get("https://piston-meta.mojang.com/mc/game/version_manifest_v2.json")
	if err != nil {
		log.Fatalf("Failed to fetch manifest: %s", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %s", err)
	}

	var manifest VersionManifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %s", err)
	}

	return &manifest
}

func fetchVersionManifest(url string) *VersionMeta {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to fetch version manifest: %s", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read version manifest: %s", err)
	}

	var meta VersionMeta
	err = json.Unmarshal(data, &meta)
	if err != nil {
		log.Fatalf("Failed to parse version JSON: %s", err)
	}

	return &meta
}

func isAllowed(rules []Rule) bool {
	if len(rules) == 0 {
		return true
	}

	currentOS := runtime.GOOS
	allowed := false

	for _, rule := range rules {
		match := rule.OS.Name == "" || rule.OS.Name == currentOS
		if match {
			if rule.Action == "allow" {
				allowed = true
			} else if rule.Action == "disallow" {
				return false
			}
		}
	}

	return allowed
}