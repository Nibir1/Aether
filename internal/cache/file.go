// internal/cache/file.go
//
// A simple file-backed cache. Files are stored as:
//
//   <cacheDir>/<sha256(key)>
//
// Each file contains:
//   timestamp\n
//   raw bytes
//
// TTL is enforced on read.

package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type fileCache struct {
	dir string
	ttl time.Duration
}

func NewFile(dir string, ttl time.Duration) Cache {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	os.MkdirAll(dir, 0o755)
	return &fileCache{dir: dir, ttl: ttl}
}

func (f *fileCache) Get(key string) ([]byte, bool) {
	path := f.filePath(key)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	parts := strings.SplitN(string(data), "\n", 2)
	if len(parts) != 2 {
		return nil, false
	}

	ts, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, false
	}

	if time.Now().After(time.Unix(ts, 0).Add(f.ttl)) {
		os.Remove(path) // expired
		return nil, false
	}

	return []byte(parts[1]), true
}

func (f *fileCache) Set(key string, value []byte, ttl time.Duration) {
	if ttl <= 0 {
		ttl = f.ttl
	}

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	content := []byte(ts + "\n" + string(value))
	os.WriteFile(f.filePath(key), content, 0o644)
}

func (f *fileCache) filePath(key string) string {
	h := sha256.Sum256([]byte(key))
	return filepath.Join(f.dir, hex.EncodeToString(h[:]))
}
