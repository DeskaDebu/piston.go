package piston

import (
	"encoding/json"
	"fmt"
)

type VersionManifest struct {
	Latest   LatestVersion  `json:"latest"`
	Versions []VersionEntry `json:"versions"`
}

type LatestVersion struct {
	Release  string `json:"release"`
	Snapshot string `json:"snapshot"`
}

type VersionEntry struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	Time        string `json:"time"`
	ReleaseTime string `json:"releaseTime"`
	SHA1        string `json:"sha1"`
}

type VersionMeta struct {
	ID             string              `json:"id"`
	Arguments      VersionArguments    `json:"arguments"`
	OlderArguments string              `json:"minecraftArguments,omitempty"`
	Downloads      map[string]Download `json:"downloads"`
	Libraries      []Library           `json:"libraries"`
	AssetIndex     AssetIndex          `json:"assetIndex"`
	MainClass      string              `json:"mainClass"`
}

type VersionArguments struct {
	Game []Argument `json:"game"`
	JVM  []Argument `json:"jvm"`
}

type Argument struct {
    Value []string
    Rules []Rule
}

func (a *Argument) UnmarshalJSON(data []byte) error {
    var str string
    if err := json.Unmarshal(data, &str); err == nil {
        a.Value = []string{str}
        return nil
    }

    var strList []string
    if err := json.Unmarshal(data, &strList); err == nil {
        a.Value = strList
        return nil
    }

    var raw struct {
        Value interface{} `json:"value"`
        Rules []Rule      `json:"rules"`
    }
    if err := json.Unmarshal(data, &raw); err != nil {
        return err
    }
    a.Rules = raw.Rules

    switch v := raw.Value.(type) {
    case string:
        a.Value = []string{v}
    case []interface{}:
        var result []string
        for _, item := range v {
            if s, ok := item.(string); ok {
                result = append(result, s)
            }
        }
        a.Value = result
    default:
        return fmt.Errorf("unsupported value type for Argument")
    }

    return nil
}

type Rule struct {
	Action string `json:"action"`
	OS     struct {
		Name string `json:"name"`
	} `json:"os"`
}

type Download struct {
	URL  string `json:"url"`
	SHA1 string `json:"sha1"`
	Size int    `json:"size"`
}

type DownloadInfo struct {
	URL  string `json:"url"`
	SHA1 string `json:"sha1"`
	Size int    `json:"size,omitempty"`
}

type LibraryDownloads struct {
	Artifact    *DownloadInfo            `json:"artifact,omitempty"`
	Classifiers map[string]*DownloadInfo `json:"classifiers,omitempty"`
}

type NativeMapping struct {
	Windows string `json:"windows,omitempty"`
	Linux   string `json:"linux,omitempty"`
	Osx     string `json:"osx,omitempty"`
}

type Library struct {
	Name      string           `json:"name"`
	Downloads LibraryDownloads `json:"downloads"`
	Natives   NativeMapping    `json:"natives,omitempty"`
	Rules     []Rule           `json:"rules,omitempty"`
}

type AssetIndex struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	SHA1 string `json:"sha1"`
}

type AssetIndexFile struct {
	Objects map[string]AssetObject `json:"objects"`
}

type AssetObject struct {
	Hash string `json:"hash"`
	Size int    `json:"size"`
}

type FabricManifest struct {
	Game   []struct {
		Version string
		Stable  bool
	} `json:"game"`
}

type FabricMeta struct {
	Loader       FabricLoader       `json:"loader"`
	LauncherMeta FabricLauncherMeta `json:"launcherMeta"`
}

type FabricLoader struct {
	Separator string `json:"separator"`
	Build     uint32 `json:"build"`
	Maven     string `json:"maven"`
	Version   string `json:"version"`
	Stable    bool   `json:"stable"`
}

type FabricLauncherMeta struct {
	Version   uint32          `json:"version"`
	Libraries FabricLibraries `json:"libraries"`
	MainClass FabricMainClass `json:"mainClass"`
}

type FabricLibraries struct {
	Client []FabricLibrary `json:"client"`
	Common []FabricLibrary `json:"common"`
	Server []FabricLibrary `json:"server"`
}

type FabricLibrary struct {
	Name string `json:"name"`
	Url  string `json:"url,omitempty"`
	Sha1 string `json:"sha1,omitempty"`
	Size int    `json:"size,omitempty"`
}

type FabricMainClass struct {
	Client string `json:"client,omitempty"`
	Server string `json:"server,omitempty"`
}

func (f *FabricMainClass) UnmarshalJSON(data []byte) error {
	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		f.Client = asString
		f.Server = asString
		return nil
	}

	var asObject struct {
		Client string `json:"client"`
		Server string `json:"server"`
	}
	if err := json.Unmarshal(data, &asObject); err != nil {
		return err
	}

	f.Client = asObject.Client
	f.Server = asObject.Server
	return nil
}