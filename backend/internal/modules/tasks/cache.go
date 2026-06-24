package tasks

import "time"

// InvalidateCache clears the tasks cache to ensure real-time data is fetched.
func InvalidateCache() {
	cacheMutex.Lock()
	tasksCache = nil
	cacheExpiration = time.Now()
	cacheMutex.Unlock()
}
