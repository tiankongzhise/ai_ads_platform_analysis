package config

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

type FileLoader struct {
	defaultPath string
	localPath   string
	getenv      func(string) string
}

func NewFileLoader(defaultPath string, localPath string, getenv func(string) string) *FileLoader {
	return &FileLoader{defaultPath: defaultPath, localPath: localPath, getenv: getenv}
}

func (l *FileLoader) Load() (map[string]string, error) {
	values := map[string]string{}
	if err := mergeYAML(values, l.defaultPath, true); err != nil {
		return nil, err
	}
	if err := mergeYAML(values, l.localPath, false); err != nil {
		return nil, err
	}
	applyEnv(values, l.getenv)
	return values, nil
}

func mergeYAML(values map[string]string, path string, required bool) error {
	file, err := os.Open(path)
	if err != nil {
		if !required && errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	section := ""
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "- ") {
			continue
		}
		if !strings.HasPrefix(line, " ") && strings.HasSuffix(trimmed, ":") {
			section = strings.TrimSuffix(trimmed, ":")
			continue
		}
		if section == "" {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := section + "." + strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		values[key] = value
	}
	return scanner.Err()
}

func applyEnv(values map[string]string, getenv func(string) string) {
	for key := range values {
		envKey := "EDUADCRM_" + strings.ToUpper(strings.NewReplacer(".", "_").Replace(key))
		if value := getenv(envKey); value != "" {
			values[key] = value
		}
	}
}
