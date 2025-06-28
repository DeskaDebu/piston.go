package piston

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"
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

func fetchFabricManifest() *FabricManifest {
	resp, err := http.Get("https://meta.fabricmc.net/v2/versions")
	if err != nil {
		log.Fatalf("Failed to fetch manifest: %s", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %s", err)
	}

	var manifest FabricManifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %s", err)
	}

	return &manifest
}

func fetchFabricLoaderManifest(version string) *[]FabricMeta {
		resp, err := http.Get("https://meta.fabricmc.net/v2/versions/loader/"+version)
	if err != nil {
		log.Fatalf("Failed to fetch manifest: %s", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %s", err)
	}

	var manifest []FabricMeta
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %s", err)
	}

	return &manifest
}

func fetchFabricLoaderMeta(version string, loader string) *FabricMeta {
	var meta FabricMeta

	for _, met := range *fetchFabricLoaderManifest(version) {
		if met.Loader.Version == loader {
			meta = met
			break
		}
	}
	
	return &meta
}

func (fabricMeta FabricMeta) patchVersionManifest(meta *VersionMeta) *VersionMeta {
	meta.MainClass = fabricMeta.LauncherMeta.MainClass.Client
	meta.ID = meta.ID + "-fabric"

	return mergeLibraries(meta, fabricMeta)
}

func mergeLibraries(meta *VersionMeta, fabricMeta FabricMeta) *VersionMeta {
    // Zrób mapę fabricowych bibliotek po group:artifact (bez wersji)
    fabricLibGA := make(map[string]Library)
    for _, lib := range fabricMeta.LauncherMeta.Libraries.Common {
        ga := groupArtifact(lib.Name)
        fabricLibGA[ga] = Library{
            Name: lib.Name,
            Downloads: LibraryDownloads{
                Artifact: &DownloadInfo{
                    URL:  lib.Name + libraryPathFromName(lib.Name),
                    SHA1: lib.Sha1,
                    Size: lib.Size,
                },
                Classifiers: map[string]*DownloadInfo{},
            },
        }
    }
    for _, lib := range fabricMeta.LauncherMeta.Libraries.Client {
        ga := groupArtifact(lib.Name)
        fabricLibGA[ga] = Library{
            Name: lib.Name,
            Downloads: LibraryDownloads{
                Artifact: &DownloadInfo{
                    URL:  lib.Name + libraryPathFromName(lib.Name),
                    SHA1: lib.Sha1,
                    Size: lib.Size,
                },
                Classifiers: map[string]*DownloadInfo{},
            },
        }
    }
    // Dodaj Loader
    gaLoader := groupArtifact(fabricMeta.Loader.Maven)
    fabricLibGA[gaLoader] = Library{
        Name: fabricMeta.Loader.Maven,
        Downloads: LibraryDownloads{
            Artifact: &DownloadInfo{
                URL: "https://maven.fabricmc.net/" + libraryPathFromName(fabricMeta.Loader.Maven),
            },
            Classifiers: map[string]*DownloadInfo{},
        },
    }

    // Wynikowa lista bibliotek
    merged := []Library{}

    // Dodaj tylko vanilla, które nie mają odpowiednika fabricowego
    for _, lib := range meta.Libraries {
        ga := groupArtifact(lib.Name)
        if _, found := fabricLibGA[ga]; !found {
            merged = append(merged, lib)
        }
    }

    // Dodaj WSZYSTKIE fabricowe biblioteki
    for _, lib := range fabricLibGA {
        merged = append(merged, lib)
    }

    meta.Libraries = merged
    return meta
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

func groupArtifact(name string) string {
    parts := strings.Split(name, ":")
    if len(parts) < 2 {
        return name
    }
    // Zwróć tylko "group:artifact"
    return parts[0] + ":" + parts[1]
}