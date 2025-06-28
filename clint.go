package piston

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func downloadClientJar(meta *VersionMeta, baseDir string) {
	clientDownload, ok := meta.Downloads["client"]
	if !ok {
		log.Fatalf("No client download in manifest!")
	}

	destDir := filepath.Join(baseDir, "versions", meta.ID)
	destJar := filepath.Join(destDir, meta.ID+".jar")

	if fileExists(destJar) && sha1Matches(destJar, clientDownload.SHA1) {
		log.Printf("client.jar for %s already exists and is valid.", meta.ID)
		return
	}

	log.Printf("Downloading client.jar for %s", meta.ID)
	err := os.MkdirAll(destDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create version dir: %s", err)
	}

	resp, err := http.Get(clientDownload.URL)
	if err != nil {
		log.Fatalf("Failed to fetch client.jar: %s", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(destJar)
	if err != nil {
		log.Fatalf("Failed to create client.jar file: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatalf("Failed to write client.jar: %s", err)
	}

	versionJSONPath := filepath.Join(destDir, meta.ID+".json")

	file, err := os.Create(versionJSONPath)
	if err != nil {
		log.Fatalf("Failed to create version.json file: %s", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(meta)
	if err != nil {
		log.Fatalf("Failed to write version.json: %s", err)
	}

	log.Printf("client.jar downloaded to %s", destJar)
}

func loadVersionManifest(baseDir string, version string) (*VersionMeta, error) {
	file, err := os.Open(filepath.Join(baseDir, "versions", version, version+".json"))
	if err != nil {
		return nil, fmt.Errorf("failed to open version.json: %w", err)
	}
	defer file.Close()

	var meta VersionMeta
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&meta)
	if err != nil {
		return nil, fmt.Errorf("failed to decode version.json: %w", err)
	}

	return &meta, nil
}