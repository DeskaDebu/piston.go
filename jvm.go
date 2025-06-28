package piston

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var varPattern = regexp.MustCompile(`\$\{([a-zA-Z0-9_]+)\}`)

var ignoredArgs = map[string]bool{
	"-Dminecraft.launcher.brand=": true,
	"-Dminecraft.launcher.version=": true,
	"--quickPlayPath":        true,
	"--quickPlaySingleplayer": true,
	"--quickPlayMultiplayer":  true,
	"--quickPlayRealms":       true,
	"--width": true,
	"--height": true,
	"--demo": true,
	"--xuid": true,
}

func replaceVars(s string, vars map[string]string) string {
	return varPattern.ReplaceAllStringFunc(s, func(match string) string {
		key := varPattern.FindStringSubmatch(match)[1]
		if val, ok := vars[key]; ok {
			return val
		}
		return match
	})
}

func expandArguments(args []Argument, replacements map[string]string) []string {
	var result []string
	for _, arg := range args {
		if !isAllowed(arg.Rules) {
			continue
		}
		for i := 0; i < len(arg.Value); i++ {
			val := replaceVars(arg.Value[i], replacements)
			
			if ignoredArgs[val] {
				continue
			}

			if strings.Contains(val, "${") {
				continue
			}

			result = append(result, val)
		}
	}
	return result
}


func buildClasspath(meta *VersionMeta, baseDir string) string {
	var paths []string

	clientJar := filepath.Join(baseDir, "versions", meta.ID, meta.ID+".jar")
	paths = append(paths, clientJar)

	for _, lib := range meta.Libraries {
		if !isAllowed(lib.Rules) {
			continue
		}
		
		ok := lib.Downloads.Artifact
		if ok == nil {
			continue
		}
		
		libPath := filepath.Join(baseDir, "libraries", libraryPathFromName(lib.Name))
		paths = append(paths, libPath)
	}

	return strings.Join(paths, string(os.PathListSeparator))
}

func buildLaunchCommand(meta *VersionMeta, baseDir string, vars map[string]string, xmx uint32) []string {
	classpath := buildClasspath(meta, baseDir)
	vars["classpath"] = classpath
	vars["natives_directory"] = filepath.Join(baseDir, "natives", meta.ID)


	jvmArgs := []string{
		fmt.Sprintf("-Xmx%dm", xmx),
		"-Djava.library.path=" + vars["natives_directory"],
		"-cp", classpath,
	}

	if meta.Arguments.Game == nil || len(meta.Arguments.Game) == 0 {
		mainClass := meta.MainClass
		jvmArgs = append(jvmArgs, mainClass)

		mcArgsStr := replaceVars(meta.OlderArguments, vars)
		gameArgs := strings.Fields(mcArgsStr)

		return append(jvmArgs, gameArgs...)
	}
	
	jvmArgs = expandArguments(meta.Arguments.JVM, vars)
	jvmArgs = append([]string{fmt.Sprintf("-Xmx%dm", xmx)}, jvmArgs...)
	jvmArgs = append(jvmArgs, meta.MainClass)

	gameArgs := expandArguments(meta.Arguments.Game, vars)

	return append(jvmArgs, gameArgs...)
}