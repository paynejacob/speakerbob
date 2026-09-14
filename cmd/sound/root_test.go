package sound

import (
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunCommandsCloseStoreOnBusinessError guards against a regression where
// exitOnError called os.Exit from inside a function with a live
// defer closeStore(), leaking the badger store handle on every
// business-logic error (not just success).
func TestRunCommandsCloseStoreOnBusinessError(t *testing.T) {
	dataPath = t.TempDir()

	updateId = "does-not-exist"
	updateName = "name"
	err := runUpdate(nil, nil)
	require.Error(t, err)

	assertStoreReopenable(t, dataPath)
}

func assertStoreReopenable(t *testing.T, path string) {
	t.Helper()

	opts := badger.DefaultOptions(path)
	opts.Logger = logrus.StandardLogger()

	db, err := badger.Open(opts)
	require.NoError(t, err, "store was not closed by the failed command, directory lock still held")
	assert.NoError(t, db.Close())
}
