package piston

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func libraryPathFromName(name string) string {
	parts := strings.Split(name, ":")
	if len(parts) < 3 {
		log.Fatalf("Invalid library name: %s", name)
	}

	group := strings.ReplaceAll(parts[0], ".", "/")
	artifact := parts[1]
	version := parts[2]

	jarName := artifact + "-" + version
	if len(parts) == 4 {
		jarName += "-" + parts[3]
	}
	jarName += ".jar"

	// Użyj path.Join, nie filepath.Join — path jest dla URL-i
	return strings.ReplaceAll(filepath.Join(group, artifact, version, jarName), "\\", "/")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func sha1Matches(path string, expected string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	actual := sha1.Sum(data)
	return hex.EncodeToString(actual[:]) == expected
}

func downloadLibrary(lib Library, baseDir string) {
	if !isAllowed(lib.Rules) {
		return
	}

	artifact := lib.Downloads.Artifact
	if artifact == nil {
		log.Printf("Skipping library %s - no artifact", lib.Name)
		return
	}

	path := libraryPathFromName(lib.Name)
	dest := filepath.Join(baseDir, "libraries", path)

	if fileExists(dest) && sha1Matches(dest, artifact.SHA1) {
		log.Printf("Library %s already downloaded.", lib.Name)
		return
	}

	err := os.MkdirAll(filepath.Dir(dest), 0755)
	if err != nil {
		log.Fatalf("Failed to create directory: %s", err)
	}

	log.Printf("Downloading %s", artifact.URL)
	resp, err := http.Get(artifact.URL)
	if err != nil {
		log.Fatalf("Failed to download: %s", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		log.Fatalf("Failed to create file: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatalf("Failed to write file: %s", err)
	}

	log.Printf("Library %s downloaded to %s", lib.Name, dest)
}