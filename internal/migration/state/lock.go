package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"
)

// LockInfo records who currently holds the migration lock. Prevents two
// concurrent runs from corrupting state.json.
type LockInfo struct {
	PID         int       `json:"pid"`
	Hostname    string    `json:"hostname"`
	AcquiredAt  time.Time `json:"acquiredAt"`
	CLIMode     string    `json:"cliMode"` // "cli" or "plugin"
}

// ErrAlreadyLocked is returned when a migration run is in progress.
var ErrAlreadyLocked = errors.New("a migration run is already in progress (lock.json exists)")

// AcquireLock creates lock.json. Fails if a lock already exists, unless the
// existing lock is stale (PID no longer running). Stale locks are reclaimed.
func AcquireLock(repoRoot, mode string) error {
	if err := os.MkdirAll(MigrationDirPath(repoRoot), 0o755); err != nil {
		return fmt.Errorf("create migration dir: %w", err)
	}
	path := LockPath(repoRoot)

	if data, err := os.ReadFile(path); err == nil {
		var existing LockInfo
		if json.Unmarshal(data, &existing) == nil && pidAlive(existing.PID) {
			return fmt.Errorf("%w (held by pid %d on %s since %s)", ErrAlreadyLocked, existing.PID, existing.Hostname, existing.AcquiredAt.Format(time.RFC3339))
		}
		// Stale lock — reclaim it.
		_ = os.Remove(path)
	}

	hostname, _ := os.Hostname()
	info := LockInfo{
		PID:        os.Getpid(),
		Hostname:   hostname,
		AcquiredAt: time.Now().UTC(),
		CLIMode:    mode,
	}
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ReleaseLock removes lock.json. Idempotent.
func ReleaseLock(repoRoot string) error {
	err := os.Remove(LockPath(repoRoot))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// pidAlive returns true if a process with pid is currently running.
// On Unix, sending signal 0 is the standard "is alive?" probe.
func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
