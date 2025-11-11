package fsutils

import "os"

func TemporaryDirectory() (string, func(), error) {
	dir, err := os.MkdirTemp("", "patreon-crawler-test-")
	if err != nil {
		return "", nil, err
	}

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup, nil
}
