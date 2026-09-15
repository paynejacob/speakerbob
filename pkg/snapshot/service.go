package snapshot

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const DefaultInterval = 24 * time.Hour

// Backuper is satisfied by any store that can stream a point-in-time backup,
// e.g. pkg/store/badgerdb.Store. It is intentionally narrower than
// hotcereal/pkg/store.Store, which has no backup capability.
type Backuper interface {
	Backup(w io.Writer) (uint64, error)
}

// Service periodically writes a full database backup to Path. It is opt-in:
// a zero-value Path disables it, so RegisterRoutes/Run are safe no-ops for
// deployments that don't configure snapshotting.
type Service struct {
	Store    Backuper
	Path     string
	Interval time.Duration
}

func (s *Service) RegisterRoutes(*mux.Router) {}

func (s *Service) Run(ctx context.Context) {
	if s.Path == "" {
		return
	}

	interval := s.Interval
	if interval <= 0 {
		interval = DefaultInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.snapshot()
		}
	}
}

func (s *Service) snapshot() {
	filename := filepath.Join(s.Path, fmt.Sprintf("snapshot-%s.bak", time.Now().UTC().Format(time.RFC3339)))
	tmpFilename := filename + ".tmp"

	f, err := os.Create(tmpFilename)
	if err != nil {
		logrus.Errorf("failed to create database snapshot file %q: %s", tmpFilename, err)
		return
	}

	if _, err := s.Store.Backup(f); err != nil {
		logrus.Errorf("failed to write database snapshot: %s", err)
		_ = f.Close()
		_ = os.Remove(tmpFilename)
		return
	}

	if err := f.Sync(); err != nil {
		logrus.Errorf("failed to sync database snapshot file %q: %s", tmpFilename, err)
		_ = f.Close()
		_ = os.Remove(tmpFilename)
		return
	}

	if err := f.Close(); err != nil {
		logrus.Errorf("failed to close database snapshot file %q: %s", tmpFilename, err)
		_ = os.Remove(tmpFilename)
		return
	}

	// A completed snapshot is written under a .tmp name and renamed into
	// place only on success, so a backup failure never leaves a
	// corrupt/truncated file indistinguishable from a real snapshot.
	if err := os.Rename(tmpFilename, filename); err != nil {
		logrus.Errorf("failed to finalize database snapshot file %q: %s", filename, err)
		_ = os.Remove(tmpFilename)
	}
}
