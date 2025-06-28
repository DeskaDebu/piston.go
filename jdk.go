package piston

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func downloadJDK(baseDir string, version uint16) error {
	url := fmt.Sprintf(
		"https://api.adoptium.net/v3/binary/latest/%d/ga/windows/x64/jdk/hotspot/normal/eclipse",
		version,
	)

	jdkFolder := filepath.Join(baseDir, "java", "jdk-"+fmt.Sprint(version))
	if _, err := os.Stat(jdkFolder); err == nil {
		log.Printf("JDK %d already downloaded", version)
		return nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch JDK: %w", err)
	}
	defer resp.Body.Close()

	zipPath := filepath.Join(baseDir, "java-temp", "jdk-" + fmt.Sprint(version) + ".zip")
	err = os.MkdirAll(filepath.Dir(zipPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create java dir: %w", err)
	}

	outFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create JDK file: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save JDK: %w", err)
	}

	err = unzip(zipPath, jdkFolder)
	if err != nil {
    	return fmt.Errorf("failed to unzip JDK: %w", err)
	}

	_ = os.Remove(zipPath)

	return nil
}

func unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest) + string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		in, err := f.Open()
		if err != nil {
			return err
		}

		out, err := os.Create(fpath)
		if err != nil {
			in.Close()
			return err
		}

		_, err = io.Copy(out, in)

		in.Close()
		out.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func findJavaExecutable(jdkDir string) string {
	javaPath := filepath.Join(jdkDir, "bin", "java.exe")
	if _, err := os.Stat(javaPath); err == nil {
		return javaPath
	}

	files, err := os.ReadDir(jdkDir)
	if err != nil {
		fmt.Println("JDK directory read failed:", err)
		return ""
	}
	for _, f := range files {
		if f.IsDir() {
			javaPath := filepath.Join(jdkDir, f.Name(), "bin", "java.exe")
			if _, err := os.Stat(javaPath); err == nil {
				return javaPath
			}
		}
	}
	fmt.Println("Could not find java.exe in:", jdkDir)
	return ""
}

func pathExists(path string) bool {
    _, err := os.Stat(path)
    return err == nil
}

func requiresJDK8(minecraftVersion string) bool {
    var major, minor int
    n, err := fmt.Sscanf(minecraftVersion, "%d.%d", &major, &minor)
    if err != nil || n < 2 {
        return true
    }

    if major < 1 {
        return true
    }
    if major == 1 && minor < 12 {
        return true
    }
    return false
}