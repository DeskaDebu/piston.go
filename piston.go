package piston

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
)

type PistonLauncher struct {
	BasePath string
	JDK8e    string
	JDK21e   string
}

func CreatePistonLauncher(BasePath string) PistonLauncher {
	jdk8path := filepath.Join(BasePath, "java", "jdk-" + fmt.Sprint(8))
	jdk21path := filepath.Join(BasePath, "java", "jdk-" + fmt.Sprint(21))

	is8Installed := pathExists(jdk8path)
	is21Installed := pathExists(jdk21path)

	if !is8Installed {
		err := downloadJDK(BasePath, 8)
		if err != nil {
        	log.Printf("Failed to download JDK 8: %v", err)
    	}
	}

	if !is21Installed {
		err := downloadJDK(BasePath, 21)
		if err != nil {
        	log.Printf("Failed to download JDK 21: %v", err)
    	}
	}

	return PistonLauncher{
		BasePath: BasePath,
		JDK8e: findJavaExecutable(jdk8path),
		JDK21e: findJavaExecutable(jdk21path),
	}
}

func (launcher PistonLauncher) IsJavaInstalled(version uint16) bool {
	return pathExists(filepath.Join(launcher.BasePath, "java", "jdk-" + fmt.Sprint(version)))
}

func (launcher PistonLauncher) DownloadJDK8() {
	downloadJDK(launcher.BasePath, 8)
}

func (launcher PistonLauncher) DownloadJDK21() {
	downloadJDK(launcher.BasePath, 21)
}

func (launcher PistonLauncher) QueryVersions() *VersionManifest {
	return fetchManifest()
}

func (launcher PistonLauncher) DownloadVersion(url string) *VersionMeta {
	meta := fetchVersionManifest(url)

	downloadClientJar(meta, launcher.BasePath)
	for _, lib := range meta.Libraries {
		downloadLibrary(lib, launcher.BasePath)
	}
	downloadNatives(meta, launcher.BasePath)
	downloadAssets(meta, launcher.BasePath)

	return meta
}

func (launcher PistonLauncher) DownloadFabricVersion(url string, version string, loader string) *VersionMeta {
	meta := fetchVersionManifest(url)
	fabricMeta := fetchFabricLoaderMeta(version, loader)

	if fabricMeta == nil {
		log.Fatalf("Couldn't find %s with loader %s.", version, loader)
	}

	meta = fabricMeta.patchVersionManifest(meta)

	downloadClientJar(meta, launcher.BasePath)
	for _, lib := range meta.Libraries {
		downloadLibrary(lib, launcher.BasePath)
	}
	downloadNatives(meta, launcher.BasePath)
	downloadAssets(meta, launcher.BasePath)

	return meta
}

func (launcher PistonLauncher) LaunchVersion(version string, xmx uint32, username string, accessToken string, uuid string, userType string, clientId string, versionType string) {
	meta, err := loadVersionManifest(launcher.BasePath, version)
	if err != nil {
		log.Fatalf("Failed to load version: %s", err)
	}

	vars := map[string]string{
		"auth_player_name":    username,
		"version_name":        version,
		"game_directory":      launcher.BasePath,
		"assets_root":         filepath.Join(launcher.BasePath, "assets"),
		"game_assets":         filepath.Join(launcher.BasePath, "assets"),
		"assets_index_name":   meta.AssetIndex.ID,
		"auth_access_token":   accessToken,
		"auth_uuid":           uuid,
		"user_type":           userType,
		"clientid":            clientId,
		"version_type":        "piston.go-" + versionType,
		"user_properties":      "{}",
	}

	log.Println("Launching Minecraft...")

	args := buildLaunchCommand(meta, launcher.BasePath, vars, xmx)

	isOlder := requiresJDK8(version)

	var jdk string

	if isOlder {
		jdk = launcher.JDK8e
	} else {
		jdk = launcher.JDK21e
	}

	cmd := exec.Command(jdk, args...)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	cmd.Stdin = nil

	err = cmd.Run()
	if err != nil {
		log.Fatalf("Minecraft process failed: %s", err)
	}
}

func (launcher PistonLauncher) GenerateOfflineUUID(username string) string {
	data := []byte("OfflinePlayer:" + username)
	hash := md5.Sum(data)

	// Format as UUID v3 (name-based UUID, MD5)
	hash[6] = (hash[6] & 0x0f) | 0x30
	hash[8] = (hash[8] & 0x3f) | 0x80

	uuid := make([]byte, 36)
	hex.Encode(uuid[0:8], hash[0:4])
	uuid[8] = '-'
	hex.Encode(uuid[9:13], hash[4:6])
	uuid[13] = '-'
	hex.Encode(uuid[14:18], hash[6:8])
	uuid[18] = '-'
	hex.Encode(uuid[19:23], hash[8:10])
	uuid[23] = '-'
	hex.Encode(uuid[24:], hash[10:])

	return string(uuid)
}