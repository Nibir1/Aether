// internal/cache/ttl.go
package cache

import "time"

func expired(created time.Time, ttl time.Duration) bool {
	if ttl <= 0 {
		return false
	}
	return time.Now().After(created.Add(ttl))
}
