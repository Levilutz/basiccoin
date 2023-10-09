package disksyncmap

import (
	"errors"
	"fmt"
	"os"

	"github.com/levilutz/basiccoin/pkg/syncmap"
)

type comparableStringer interface {
	comparable
	fmt.Stringer
}

// A sync map backed by disk.
type DiskSyncMap[K comparableStringer, V fmt.Stringer] struct {
	m         *syncmap.SyncMap[K, V]
	basePath  string
	parseFunc func(raw string) (v V, err error)
}

// Create a new disk-backed sync map.
func NewDiskSyncMap[K comparableStringer, V fmt.Stringer](
	basePath string,
	parseFunc func(raw string) (v V, err error),
) *DiskSyncMap[K, V] {
	if err := os.MkdirAll(basePath, 0750); err != nil {
		panic(fmt.Sprintf("failed to make DiskSyncMap dir: %s", err))
	}
	return &DiskSyncMap[K, V]{
		m:         syncmap.NewSyncMap[K, V](),
		basePath:  basePath,
		parseFunc: parseFunc,
	}
}

// Check whether the given key exists in the disk-backend sync map.
func (dsm *DiskSyncMap[K, V]) Has(key K) bool {
	if dsm.m.Has(key) {
		return true
	}
	return dsm.tryLoadKeyFromDisk(key)
}

// Retrieve the given key from the disk map.
func (dsm *DiskSyncMap[K, V]) Get(key K) V {
	exists := dsm.tryLoadKeyFromDisk(key)
	if !exists {
		panic(fmt.Sprintf("attempted to load unknown key: %s", key))
	}
	return dsm.m.Get(key)
}

// Store the given key to the map and disk.
func (dsm *DiskSyncMap[K, V]) Store(key K, val V) {
	dsm.m.Store(key, val)
	if err := dsm.saveFile(dsm.keyPath(key), val); err != nil {
		fmt.Printf("failed to save file %s: %s\n", dsm.keyPath(key), err)
	}
}

// Load the key from files into the sync map if not present, return whether it exists.
func (dsm *DiskSyncMap[K, V]) tryLoadKeyFromDisk(key K) bool {
	if dsm.m.Has(key) {
		return true
	}
	val, err := dsm.loadFile(dsm.keyPath(key))
	if err != nil {
		return false
	}
	dsm.m.Store(key, val)
	return true
}

// Write to the given file on disk.
func (dsm *DiskSyncMap[K, V]) saveFile(path string, val V) error {
	return os.WriteFile(path, []byte(val.String()), 0666)
}

// Load the given file from disk and parse.
func (dsm *DiskSyncMap[K, V]) loadFile(path string) (val V, err error) {
	if !dsm.fileExists(path) {
		return val, fmt.Errorf("file does not exist: %s", path)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return val, fmt.Errorf("failed to read %s: %s", path, err)
	}
	val, err = dsm.parseFunc(string(raw))
	if err != nil {
		return val, fmt.Errorf("failed to parse %s: %s", path, err)
	}
	return val, nil
}

// Check whether the given file exists on disk.
func (dsm *DiskSyncMap[K, V]) fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		fmt.Printf("failed to check existence of file %s: %s\n", path, err)
		return false
	}
}

func (dsm *DiskSyncMap[K, V]) keyPath(key K) string {
	return dsm.basePath + "/" + key.String()
}
