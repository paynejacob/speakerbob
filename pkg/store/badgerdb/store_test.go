package badgerdb

import (
	"bytes"
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/paynejacob/hotcereal/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openTestStore(t *testing.T) Store {
	t.Helper()

	opts := badger.DefaultOptions(t.TempDir()).WithLoggingLevel(badger.WARNING)
	db, err := badger.Open(opts)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	return Store{DB: db}
}

func TestStoreBackupContainsSavedData(t *testing.T) {
	s := openTestStore(t)

	key := store.TypeKey{Body: "versionVersion", PackageLength: 7, TypeLength: 7}
	require.NoError(t, s.Save(key, []byte("hello")))

	var buf bytes.Buffer
	n, err := s.Backup(&buf)

	require.NoError(t, err)
	assert.Greater(t, n, uint64(0))
	assert.NotEmpty(t, buf.Bytes())
}
