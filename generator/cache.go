package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type IDLCache struct {
	Cache map[string]*cache

	//ensure each remote repo is only cloned once
	RemoteRepoFlag map[string]bool
}

type cache struct {
	Hashes    string
	TimeStamp time.Time
}

type IDLStatus uint

func NewIDLCache() *IDLCache {
	return &IDLCache{
		Cache:          make(map[string]*cache),
		RemoteRepoFlag: make(map[string]bool),
	}
}

func (ic *IDLCache) LoadCache(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &ic.Cache)
}

func (ic *IDLCache) SaveCache(filePath string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.Marshal(ic.Cache)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func (ic *IDLCache) CalculateHash(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func (ic *IDLCache) AddHash(filePath string) error {
	newHash, err := ic.CalculateHash(filePath)
	if err != nil {
		return err
	}

	ic.Cache[filePath] = &cache{
		Hashes:    newHash,
		TimeStamp: time.Now(),
	}

	return nil
}

func (ic *IDLCache) HashHasChanged(filePath string) (bool, error) {
	newHash, err := ic.CalculateHash(filePath)
	if err != nil {
		return false, err
	}
	c, exists := ic.Cache[filePath]
	if !exists || c.Hashes != newHash {
		ic.Cache[filePath] = &cache{
			Hashes:    newHash,
			TimeStamp: time.Now(),
		}
		return true, nil
	}
	return false, nil
}
