package i18n

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	dict      = make(map[string]map[string]string) // dict[lang][key] = value
	dictMutex sync.RWMutex
)

func LoadDirectoryFromEnv() error {
	path := os.Getenv("I18N_PATH")
	if path == "" {
		return nil
	}
	return LoadDirectory(path)
}

// LoadDirectory load toàn bộ thư mục i18n
func LoadDirectory(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	tmp := make(map[string]map[string]string)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		lang := entry.Name()
		langPath := filepath.Join(root, lang)

		tmp[lang] = make(map[string]string)

		// Đọc tất cả file JSON trong thư mục lang
		err := filepath.WalkDir(langPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			if filepath.Ext(path) != ".json" {
				return nil
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			m := make(map[string]string)
			if err := json.Unmarshal(data, &m); err != nil {
				return err
			}

			for k, v := range m {
				k = strings.ToLower(strings.TrimSpace(k))
				k = strings.TrimSuffix(k, ".") // remove trailing dot
				tmp[lang][k] = v
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	dictMutex.Lock()
	dict = tmp
	dictMutex.Unlock()
	return nil
}

// Translate global
func Translate(content, lang string) string {
	if lang == "" {
		return content
	}
	key := strings.ToLower(strings.TrimSpace(content))
	key = strings.TrimSuffix(key, ".") // remove trailing dot

	if lang == "" {
		return key
	}

	dictMutex.RLock()
	m, ok := dict[lang]
	dictMutex.RUnlock()

	if ok {
		if val, ok2 := m[key]; ok2 && val != "" {
			return val
		}
	}

	// fallback en
	dictMutex.RLock()
	fallback, ok := dict["en"][key]
	dictMutex.RUnlock()
	if ok && fallback != "" {
		return fallback
	}

	return content
}
