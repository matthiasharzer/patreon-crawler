package fsutils_test

import (
	"os"
	"path/filepath"
	"testing"

	"patreon-crawler/util/fsutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemporaryDirectory(t *testing.T) {
	t.Run("creates and cleans temporary directory", func(t *testing.T) {
		dir, cleanup, err := fsutils.TemporaryDirectory()
		require.NoError(t, err)
		require.NotEmpty(t, dir)
		require.NotNil(t, cleanup)

		info, err := os.Stat(dir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())

		fpath := filepath.Join(dir, "testfile.txt")
		err = os.WriteFile(fpath, []byte("hello"), 0644)
		require.NoError(t, err)

		_, err = os.Stat(fpath)
		require.NoError(t, err)

		cleanup()

		_, err = os.Stat(dir)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("cleanup can be called multiple times", func(t *testing.T) {
		dir, cleanup, err := fsutils.TemporaryDirectory()
		require.NoError(t, err)
		require.NotEmpty(t, dir)
		require.NotNil(t, cleanup)

		cleanup()
		cleanup()

		_, err = os.Stat(dir)
		assert.True(t, os.IsNotExist(err))
	})
}
