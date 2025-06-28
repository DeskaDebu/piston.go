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

func downloadAsset(obj AssetObject, baseDir string) {
	prefix := obj.Hash[:2]
	url := fmt.Sprintf("https://resources.download.minecraft.net/%s/%s", prefix, obj.Hash)
	dest := filepath.Join(baseDir, "assets", "objects", prefix, obj.Hash)

	if fileExists(dest) && sha1Matches(dest, obj.Hash) {
		return
	}

	err := os.MkdirAll(filepath.Dir(dest), 0755)
	if err != nil {
		log.Fatalf("Failed to create asset directory: %s", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to fetch asset: %s", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		log.Fatalf("Failed to create asset file: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatalf("Failed to write asset: %s", err)
	}
}

func downloadAssets(meta *VersionMeta, baseDir string) {
    log.Printf("Fetching asset index: %s", meta.AssetIndex.URL)
    resp, err := http.Get(meta.AssetIndex.URL)
    if err != nil {
        log.Fatalf("Failed to fetch asset index: %s", err)
    }
    defer resp.Body.Close()

    data, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Fatalf("Failed to read asset index: %s", err)
    }

    indexPath := filepath.Join(baseDir, "assets", "indexes", meta.AssetIndex.ID+".json")
    err = os.MkdirAll(filepath.Dir(indexPath), 0755)
    if err != nil {
        log.Fatalf("Failed to create indexes directory: %s", err)
    }
    err = os.WriteFile(indexPath, data, 0644)
    if err != nil {
        log.Fatalf("Failed to write asset index file: %s", err)
    }

    var index AssetIndexFile
    err = json.Unmarshal(data, &index)
    if err != nil {
        log.Fatalf("Failed to parse asset index: %s", err)
    }

    count := 0
    for _, obj := range index.Objects {
        downloadAsset(obj, baseDir)
        count++
        if count%100 == 0 {
            log.Printf("Downloaded %d assets...", count)
        }
    }

    log.Printf("All assets downloaded (%d files)", count)
}
