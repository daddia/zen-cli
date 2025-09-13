package cmdutil

import (
	"errors"
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestFactory(t *testing.T) {
	t.Run("creates factory with basic fields", func(t *testing.T) {
		f := &Factory{
			AppVersion:     "1.0.0",
			ExecutableName: "zen",
			IOStreams:      iostreams.Test(),
			Logger:         logging.NewBasic(),
		}

		assert.Equal(t, "1.0.0", f.AppVersion)
		assert.Equal(t, "zen", f.ExecutableName)
		assert.NotNil(t, f.IOStreams)
		assert.NotNil(t, f.Logger)
	})

	t.Run("config function returns config", func(t *testing.T) {
		expectedConfig := &config.Config{
			LogLevel: "info",
		}

		f := &Factory{
			Config: func() (*config.Config, error) {
				return expectedConfig, nil
			},
		}

		cfg, err := f.Config()
		assert.NoError(t, err)
		assert.Equal(t, expectedConfig, cfg)
	})

	t.Run("config function returns error", func(t *testing.T) {
		expectedErr := errors.New("config error")

		f := &Factory{
			Config: func() (*config.Config, error) {
				return nil, expectedErr
			},
		}

		cfg, err := f.Config()
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Equal(t, expectedErr, err)
	})
}
