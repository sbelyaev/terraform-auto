package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/namsral/flag"
)

var (
	flagDebug     = flag.Bool("debug", false, "enable dubug output")
	defaultFilter = flag.String("default-filter", "= 0.13.6", "set default filter to '= 0.13.6'")
	osPath        = flag.String("path", "", "system path variable")
)

/*
 Usage:
 1) put required terraform binaries to valid path specified by $PATH in form "terraform-x.x.x", e.g.
   /usr/local/bin/terraform-0.11.14
   /usr/local/bin/terraform-0.13.6
   /usr/local/bin/terraform-1.0.4
 2) call this tool from directory with .tf configs
 3a) the tool loads "required_version" from config, e.g.
   required_version = "~> 1.0.4"
 3b) default behavior is '= 0.13.6' if the tool can't load a version filter
 4) this tool will call a suitable binary depend on filter applied
*/
func main() {
	flag.Parse()
	tfmBins := InitTfmBins(osPath)
	for verStr, file := range tfmBins {
		fmt.Printf("version: %s, file: %s\n", verStr, file)
	}
}

// InitTfmBins locates terraform binaries through $PATH
func InitTfmBins(osPath *string) map[string]string {
	tfmBins := make(map[string]string)
	pathDelimeter := ":"
	tfmRegexStr := `^(.*)/terraform-(.*)$`
	if runtime.GOOS == "windows" {
		pathDelimeter = ";"
		tfmRegexStr = `^(.*)\\terraform-(.*)\.exe$`
	}
	execPaths := strings.Split(*osPath, pathDelimeter)
	myDebug("Paths to locate terraform: %v", execPaths)
	for i, oneDir := range execPaths {
		myDebug("%d) %s", i, oneDir)
		if f, err := os.Open(oneDir); err == nil {
			matches, _ := filepath.Glob(fmt.Sprintf("%s/terraform-*", oneDir))
			for _, match := range matches {
				if fileInfo, err := os.Lstat(match); err == nil {
					if fileInfo.Mode().IsRegular() && (fileInfo.Mode().Perm()&0111 == 0111) {
						tfmVersionRegex := regexp.MustCompile(tfmRegexStr)
						tfmVersion := tfmVersionRegex.ReplaceAllString(match, "$2")
						if tfmBins[tfmVersion] == "" {
							tfmBins[tfmVersion] = match
						}
					}
				}
			}
			f.Close()
		}
	}
	if *flagDebug {
		for verStr, file := range tfmBins {
			myDebug("version: %s, file: %s", verStr, file)
		}
	}
	return tfmBins
}

// Print if debug is enabled
func myDebug(format string, a ...interface{}) {
	if *flagDebug {
		format = fmt.Sprintf("[DEBUG] %s\n", format)
		fmt.Fprintf(os.Stderr, format, a...)
	}
}
