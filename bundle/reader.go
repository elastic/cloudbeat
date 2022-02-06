package bundle

import (
	"io/fs"
	"os"
	"strings"
)

var includeFileSuffixes = []string{
	"data.json",
	"data.yaml",
	"data.yml",
	".rego",
}

var excludeFileSuffixes = []string{
	"test.rego",
}

func createPolicyMap(fsys fs.FS, filePrefixes []string) (map[string]string, error) {
	policies := make(map[string]string)

	err := fs.WalkDir(fsys, ".", func(filePath string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		include := !info.IsDir() && includeFile(filePrefixes, filePath)
		if !include {
			return nil
		}

		data, err := fs.ReadFile(fsys, filePath)
		if err != nil {
			return err
		}

		policies[filePath] = string(data)
		return nil
	})

	return policies, err
}

func includeFile(filePrefixes []string, filePath string) bool {
	return hasPrefix(filePath, filePrefixes) &&
		hasSuffix(filePath, includeFileSuffixes) &&
		!hasSuffix(filePath, excludeFileSuffixes)
}

func hasPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		matched := strings.HasPrefix(s, p)
		if matched {
			return true
		}
	}

	return false
}

func hasSuffix(s string, suffixes []string) bool {
	for _, p := range suffixes {
		matched := strings.HasSuffix(s, p)
		if matched {
			return true
		}
	}

	return false
}
