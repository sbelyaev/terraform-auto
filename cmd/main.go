package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"io/ioutil"

	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

var (
	flagDebug         bool
	defaultConstraint string
	osPath            string
)

func InitVars() {
	v, exists := os.LookupEnv("PATH")
	if exists {
		osPath = v
	}
	v, exists = os.LookupEnv("DEBUG")
	if exists && v == "true" {
		flagDebug = true
	} else {
		flagDebug = false
	}
	v, exists = os.LookupEnv("DEFAULT_CONSTRAINT")
	if exists {
		defaultConstraint = v
	} else {
		defaultConstraint = "= 0.13.6"
	}
}

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
	InitVars()
	tfmBins := InitTfmBins(&osPath)
	err, verConstraints := ParseTfConfigs("./")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %s", err.Error())
		os.Exit(1)
	}
	tfmBinFile := SelectTfmBin(verConstraints, tfmBins)
	myDebug("Calling: %s", tfmBinFile)
	if tfmBinFile == "" {
		fmt.Println("[ERROR] there is no file to execute")
		os.Exit(1)
	}
	out, err := exec.Command(tfmBinFile, os.Args[1:]...).CombinedOutput()
	fmt.Println(string(out[:]))
	if err != nil {
		os.Exit(1)
	}
}

// SelectTfmBin searches suitable binary to execute
func SelectTfmBin(c string, b map[string]string) string {
	tfVerConstraints, _ := version.NewConstraint(c)
	var (
		finalVersionStr string
		finaFile        string
	)
	for verStr, file := range b {
		myDebug("(%s): %s", verStr, file)
		tfNewVersion, _ := version.NewVersion(verStr)
		myDebug("%v vs %v", tfNewVersion, tfVerConstraints)
		if tfVerConstraints.Check(tfNewVersion) {
			myDebug("Got terraform bin: %s", file)
			if finalVersionStr != "" {
				oldVersion, _ := version.NewVersion(finalVersionStr)
				if oldVersion.LessThan(tfNewVersion) {
					finalVersionStr = verStr
					finaFile = file
				}
			} else {
				finalVersionStr = verStr
				finaFile = file
			}
		} else {
			myDebug("Skip file: %s", file)
		}
	}
	return finaFile
}

// ParseTfConfigs reads *.tf files to get value of "required_version" attribute
func ParseTfConfigs(workdirDir string) (error, string) {
	var versionString string
	configRootSchema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type: "terraform",
			},
		},
	}
	configFileVersionSniffBlockSchema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: "required_version",
			},
		},
	}
	matches, _ := filepath.Glob(fmt.Sprintf("%s/*.tf", workdirDir))
	for _, path := range matches {
		if src, err := ioutil.ReadFile(path); err == nil {
			file, _ := hclparse.NewParser().ParseHCL(src, path)
			if file == nil || file.Body == nil {
				continue
			}
			rootContent, _, _ := file.Body.PartialContent(configRootSchema)
			for _, block := range rootContent.Blocks {
				content, _, _ := block.Body.PartialContent(configFileVersionSniffBlockSchema)
				attr, exists := content.Attributes["required_version"]
				if !exists || len(versionString) > 0 {
					continue
				}
				val, diags := attr.Expr.Value(nil)
				if diags.HasErrors() {
					err = fmt.Errorf("Error in attribute value")
					return err, ""
				}
				versionString = val.AsString()
			}
		}
	}
	if versionString == "" {
		versionString = defaultConstraint
	}
	return nil, versionString
}

// InitTfmBins locates terraform binaries through $PATH
func InitTfmBins(osPath *string) map[string]string {
	tfmBins := make(map[string]string)
	pathDelimeter := ":"
	tfmRegexStr := `^(.*)/terraform-([0-9]*\.[0-9]*\.[0-9]*)$`
	if runtime.GOOS == "windows" {
		pathDelimeter = ";"
		tfmRegexStr = `^(.*)\\terraform-([0-9]*\.[0-9]*\.[0-9]*)\.exe$`
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
						if tfmVersionRegex.MatchString(match) {
							tfmVersion := tfmVersionRegex.ReplaceAllString(match, "$2")
							if tfmBins[tfmVersion] == "" {
								tfmBins[tfmVersion] = match
							}
						}
					}
				}
			}
			f.Close()
		}
	}
	if flagDebug {
		for verStr, file := range tfmBins {
			myDebug("version: %s, file: %s", verStr, file)
		}
	}
	return tfmBins
}

// myDebug prints to stderr if debug is enabled
func myDebug(format string, a ...interface{}) {
	if flagDebug {
		format = fmt.Sprintf("[DEBUG] %s\n", format)
		fmt.Fprintf(os.Stderr, format, a...)
	}
}
