package piston

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func extractNatives(jarPath string, outputDir string) {
	reader, err := zip.OpenReader(jarPath)
	if err != nil {
		log.Fatalf("Failed to open native jar: %s", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "META-INF") || strings.HasSuffix(file.Name, ".class") {
			continue
		}

		if !strings.HasSuffix(file.Name, ".so") &&
			!strings.HasSuffix(file.Name, ".dll") &&
			!strings.HasSuffix(file.Name, ".dylib") &&
			!strings.HasSuffix(file.Name, ".jnilib") {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			log.Fatalf("Failed to open file in jar: %s", err)
		}
		defer rc.Close()

		dest := filepath.Join(outputDir, filepath.Base(file.Name))
		out, err := os.Create(dest)
		if err != nil {
			log.Fatalf("Failed to create extracted native: %s", err)
		}
		defer out.Close()

		_, err = io.Copy(out, rc)
		if err != nil {
			log.Fatalf("Failed to copy native file: %s", err)
		}
	}
}

func downloadNatives(meta *VersionMeta, baseDir string) {
	currentOS := runtime.GOOS
	outputDir := filepath.Join(baseDir, "natives", meta.ID)
	_ = os.MkdirAll(outputDir, 0755)

	for _, lib := range meta.Libraries {
		if !isAllowed(lib.Rules) {
			continue
		}
		key := ""
		switch currentOS {
		case "windows":
			key = lib.Natives.Windows
		case "linux":
			key = lib.Natives.Linux
		case "darwin":
			key = lib.Natives.Osx
		}
		if key == "" {
			continue
		}

		classifier := lib.Downloads.Classifiers
		if classifier == nil {
			continue
		}

		nativeDownload, ok := classifier[key]
		if !ok {
			continue
		}

		dest := filepath.Join(baseDir, "libraries", "natives", filepath.Base(nativeDownload.URL))
		if fileExists(dest) && sha1Matches(dest, nativeDownload.SHA1) {
			log.Printf("Native %s already downloaded", lib.Name)
		} else {
			resp, err := http.Get(nativeDownload.URL)
			if err != nil {
				log.Fatalf("Failed to download native: %s", err)
			}
			defer resp.Body.Close()

			os.MkdirAll(filepath.Dir(dest), 0755)
			out, _ := os.Create(dest)
			io.Copy(out, resp.Body)
			out.Close()
		}

		extractNatives(dest, outputDir)
	}
}