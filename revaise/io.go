package revaise

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func loadRecord(path string) (Record, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Record{}, false, nil
		}
		return nil, false, err
	}
	if len(data) == 0 {
		return Record{}, true, nil
	}

	record := Record{}
	switch detectFormat(path, "") {
	case "yaml":
		if err := yaml.Unmarshal(data, &record); err != nil {
			return nil, true, err
		}
	default:
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, true, err
		}
	}
	return record, true, nil
}

func saveRecord(path, format string, record Record) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil && filepath.Dir(path) != "." {
		return err
	}

	var data []byte
	var err error
	switch detectFormat(path, format) {
	case "yaml":
		data, err = yaml.Marshal(record)
	default:
		data, err = json.MarshalIndent(record, "", "  ")
		if err == nil {
			data = append(data, '\n')
		}
	}
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}

func backupRecord(cfg Config) (string, error) {
	if !cfg.backupEnabled() {
		return "", nil
	}
	data, err := os.ReadFile(cfg.RecordFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	recordDir := filepath.Dir(cfg.RecordFile)
	backupDir := cfg.BackupDir
	if backupDir == "" {
		backupDir = filepath.Join(recordDir, ".revaise-backups")
	}
	if !filepath.IsAbs(backupDir) {
		backupDir = filepath.Join(recordDir, backupDir)
	}
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}

	base := filepath.Base(cfg.RecordFile)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	if ext == "" {
		ext = "." + detectFormat(cfg.RecordFile, cfg.Format)
	}
	backupName := fmt.Sprintf("%s.%s.bak%s", stem, time.Now().UTC().Format("20060102T150405.000000000Z"), ext)
	backupPath := filepath.Join(backupDir, backupName)
	return backupPath, os.WriteFile(backupPath, data, 0644)
}

func detectFormat(path, explicit string) string {
	switch strings.ToLower(strings.TrimSpace(explicit)) {
	case "yaml", "yml":
		return "yaml"
	case "json":
		return "json"
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		return "yaml"
	default:
		return "json"
	}
}

// Update loads a RevAIse record, applies a change, backs up the previous file
// when present, and saves the updated record atomically.
func Update(cfg Config, seed ReviewSeed, apply func(Record) error) error {
	if !cfg.IsEnabled() {
		return nil
	}
	if err := requireEnabled(cfg); err != nil {
		return err
	}

	record, existed, err := loadRecord(cfg.RecordFile)
	if err != nil {
		return err
	}
	ensureRoot(record, cfg, seed)
	if err := apply(record); err != nil {
		return err
	}
	record["updated_at"] = nowRFC3339()

	if existed {
		if _, err := backupRecord(cfg); err != nil {
			return err
		}
	}
	return saveRecord(cfg.RecordFile, cfg.Format, record)
}
