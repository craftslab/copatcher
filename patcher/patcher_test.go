package patcher

import (
	"context"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPatcher(t *testing.T) {
	p := New(context.Background(), DefaultConfig())
	assert.NotEqual(t, nil, p)
}

func TestRemoveIfNotDebug(t *testing.T) {
	t.Run("RemoveWorkingFolder", func(t *testing.T) {
		log.SetLevel(log.InfoLevel)

		workingFolder := t.TempDir()
		defer func(path string) {
			_ = os.RemoveAll(path)
		}(workingFolder)

		removeIfNotDebug(workingFolder)

		if _, err := os.Stat(workingFolder); err == nil {
			t.Errorf("Working folder should have been removed but still exists")
		}
	})

	t.Run("KeepWorkingFolderDebug", func(t *testing.T) {
		log.SetLevel(log.DebugLevel)

		workingFolder := t.TempDir()

		removeIfNotDebug(workingFolder)

		if _, err := os.Stat(workingFolder); err != nil {
			t.Errorf("Working folder should have been kept but was removed")
		}

		_ = os.RemoveAll(workingFolder)
	})
}
