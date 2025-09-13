package init

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/jonathandaddia/zen/pkg/cmdutil"
	"github.com/jonathandaddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdInit(t *testing.T) {
	f := &cmdutil.Factory{
		IOStreams: iostreams.Test(),
	}

	cmd := NewCmdInit(f)

	assert.Equal(t, "init", cmd.Use)
	assert.Equal(t, "Initialize a new Zen workspace", cmd.Short)
	assert.NotNil(t, cmd.RunE)

	// Check flags
	assert.NotNil(t, cmd.Flags().Lookup("force"))
	assert.NotNil(t, cmd.Flags().Lookup("config"))
}

func TestInitCommand(t *testing.T) {
	t.Run("creates new config file", func(t *testing.T) {
		// Create temp directory
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "zen.yaml")

		// Change to temp directory
		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		out := &bytes.Buffer{}
		f := &cmdutil.Factory{
			IOStreams: &iostreams.IOStreams{
				Out: out,
			},
		}

		cmd := NewCmdInit(f)
		err := cmd.Execute()
		require.NoError(t, err)

		// Check file was created
		_, err = os.Stat(configFile)
		assert.NoError(t, err)

		// Check output
		output := out.String()
		assert.Contains(t, output, "✅ Zen workspace initialized successfully!")
		assert.Contains(t, output, configFile)
	})

	t.Run("fails if config exists without force", func(t *testing.T) {
		// Create temp directory with existing config
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "zen.yaml")
		os.WriteFile(configFile, []byte("existing"), 0644)

		// Change to temp directory
		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		f := &cmdutil.Factory{
			IOStreams: iostreams.Test(),
		}

		cmd := NewCmdInit(f)
		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("overwrites with force flag", func(t *testing.T) {
		// Create temp directory with existing config
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "zen.yaml")
		os.WriteFile(configFile, []byte("existing"), 0644)

		// Change to temp directory
		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		out := &bytes.Buffer{}
		f := &cmdutil.Factory{
			IOStreams: &iostreams.IOStreams{
				Out: out,
			},
		}

		cmd := NewCmdInit(f)
		cmd.SetArgs([]string{"--force"})
		err := cmd.Execute()
		require.NoError(t, err)

		// Check file was overwritten
		content, _ := os.ReadFile(configFile)
		assert.NotEqual(t, "existing", string(content))
		assert.Contains(t, string(content), "version:")
	})
}
