package snapshot

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeBackuper struct {
	calls int
}

func (f *fakeBackuper) Backup(w io.Writer) (uint64, error) {
	f.calls++
	n, err := w.Write([]byte("backup"))
	return uint64(n), err
}

type failingBackuper struct{}

func (failingBackuper) Backup(w io.Writer) (uint64, error) {
	_, _ = w.Write([]byte("partial"))
	return 0, errors.New("simulated backup failure")
}

func TestServiceRunWritesSnapshotOnInterval(t *testing.T) {
	dir := t.TempDir()
	backuper := &fakeBackuper{}

	svc := &Service{Store: backuper, Path: dir, Interval: 10 * time.Millisecond}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		svc.Run(ctx)
		close(done)
	}()

	assert.Eventually(t, func() bool {
		entries, err := os.ReadDir(dir)
		return err == nil && len(entries) > 0
	}, time.Second, 5*time.Millisecond)

	cancel()
	<-done

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	contents, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	require.NoError(t, err)
	assert.Equal(t, "backup", string(contents))
}

func TestSnapshotLeavesNoFileOnBackupFailure(t *testing.T) {
	dir := t.TempDir()
	svc := &Service{Store: failingBackuper{}, Path: dir}

	svc.snapshot()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		assert.Falsef(t, strings.HasSuffix(e.Name(), ".bak"), "expected no completed snapshot file after a failed backup, found %q", e.Name())
		assert.Falsef(t, strings.HasSuffix(e.Name(), ".tmp"), "expected the temp file to be cleaned up after a failed backup, found %q", e.Name())
	}
}

func TestServiceRunNoopWhenPathUnset(t *testing.T) {
	backuper := &fakeBackuper{}
	svc := &Service{Store: backuper, Interval: time.Millisecond}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		svc.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Run did not return promptly for an unconfigured (empty Path) service")
	}
	cancel()

	assert.Equal(t, 0, backuper.calls)
}
