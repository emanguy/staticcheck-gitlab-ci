package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type StaticCheckEntry struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Location struct {
		File   string `json:"file"`
		Line   int    `json:"line"`
		Column int    `json:"column"`
	} `json:"location"`
	End     interface{} `json:"end"`
	Message string      `json:"message"`
}

type GitlabCIEntry struct {
	CheckName   string `json:"check_name"`
	Description string `json:"description"`
	Fingerprint string `json:"fingerprint"`
	Severity    string `json:"severity"`
	Location    struct {
		Path  string `json:"path"`
		Lines struct {
			Begin int `json:"begin"`
		} `json:"lines"`
	} `json:"location"`
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	var gitlabEntries = make([]GitlabCIEntry, 0)
	for scanner.Scan() {
		var entry StaticCheckEntry
		err := json.Unmarshal([]byte(scanner.Text()), &entry)
		if err != nil {
			log.Fatal(err)
		}

		var gitlabEntry GitlabCIEntry
		gitlabEntry.CheckName = entry.Code
		gitlabEntry.Description = entry.Message
		gitlabEntry.Fingerprint = fmt.Sprintf("%s%s%d%d", entry.Code, entry.Location.File, entry.Location.Line, entry.Location.Column)
		gitlabEntry.Severity = staticcheckSevToGitlabSev(entry.Severity)

		gitlabEntry.Location.Path = getRelativePath(entry.Location.File)
		gitlabEntry.Location.Lines.Begin = entry.Location.Line
		gitlabEntries = append(gitlabEntries, gitlabEntry)
	}
	gitlabJson, err := json.Marshal(gitlabEntries)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
	fmt.Println(string(gitlabJson))
	if len(gitlabEntries) == 0 {
		os.Exit(0)
	}
	os.Exit(1)
}

func getRelativePath(absolutePath string) string {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return strings.ReplaceAll(absolutePath, path+"/", "")
}

func staticcheckSevToGitlabSev(scSev string) string {
	// Staticcheck severities pulled from here: https://github.com/dominikh/go-tools/blob/915b568982be0ad65a98e822471748b328240ed0/lintcmd/lint.go#L373
	// Gitlab severities pulled from here: https://docs.gitlab.com/ee/ci/testing/code_quality.html#implement-a-custom-tool
	switch scSev {
	case "ignored":
		return "info"
	case "warning":
		return "minor"
	case "error":
		return "major"
	default:
		return scSev
	}
}
