package performance

import (
	"testing"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/internal/workspace"
	"github.com/stretchr/testify/require"
)

// BenchmarkConfigLoad tests the performance of configuration loading
func BenchmarkConfigLoad(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg, err := config.Load()
		if err != nil {
			b.Fatal(err)
		}
		_ = cfg
	}
}

// BenchmarkGetConfig tests the performance of getting component configuration
func BenchmarkGetConfig(b *testing.B) {
	cfg := config.LoadDefaults()
	parser := assets.ConfigParser{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		assetConfig, err := config.GetConfig(cfg, parser)
		if err != nil {
			b.Fatal(err)
		}
		_ = assetConfig
	}
}

// BenchmarkSetConfig tests the performance of setting component configuration
func BenchmarkSetConfig(b *testing.B) {
	cfg := config.LoadDefaults()
	parser := assets.ConfigParser{}
	assetConfig, err := config.GetConfig(cfg, parser)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := config.SetConfig(cfg, parser, assetConfig)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestConfigLoadPerformance tests that config loading meets P95 ≤ 10ms requirement
func TestConfigLoadPerformance(t *testing.T) {
	const iterations = 1000
	const p95Threshold = 10 * time.Millisecond
	
	durations := make([]time.Duration, iterations)
	
	for i := 0; i < iterations; i++ {
		start := time.Now()
		cfg, err := config.Load()
		duration := time.Since(start)
		
		require.NoError(t, err)
		require.NotNil(t, cfg)
		
		durations[i] = duration
	}
	
	// Calculate performance metrics
	var total time.Duration
	var max time.Duration
	for _, d := range durations {
		total += d
		if d > max {
			max = d
		}
	}
	
	avg := total / time.Duration(iterations)
	
	t.Logf("Config loading performance:")
	t.Logf("  Average: %v", avg)
	t.Logf("  Max: %v", max)
	t.Logf("  P95 threshold: %v", p95Threshold)
	
	// Use max as a proxy for P95 (conservative estimate)
	if max > p95Threshold {
		t.Errorf("Config loading performance does not meet P95 ≤ %v requirement (max: %v)", p95Threshold, max)
	}
}

// TestGetConfigPerformance tests that component config parsing meets performance requirements
func TestGetConfigPerformance(t *testing.T) {
	const iterations = 1000
	const p95Threshold = 1 * time.Millisecond
	
	cfg := config.LoadDefaults()
	parser := assets.ConfigParser{}
	durations := make([]time.Duration, iterations)
	
	for i := 0; i < iterations; i++ {
		start := time.Now()
		assetConfig, err := config.GetConfig(cfg, parser)
		duration := time.Since(start)
		
		require.NoError(t, err)
		require.NotEmpty(t, assetConfig.RepositoryURL)
		
		durations[i] = duration
	}
	
	// Calculate performance metrics
	var total time.Duration
	var max time.Duration
	for _, d := range durations {
		total += d
		if d > max {
			max = d
		}
	}
	
	avg := total / time.Duration(iterations)
	
	t.Logf("Component config parsing performance:")
	t.Logf("  Average: %v", avg)
	t.Logf("  Max: %v", max)
	t.Logf("  P95 threshold: %v", p95Threshold)
	
	// Use max as a proxy for P95 (conservative estimate)
	if max > p95Threshold {
		t.Errorf("Component config parsing does not meet P95 ≤ %v requirement (max: %v)", p95Threshold, max)
	}
}

// TestConcurrentConfigAccess tests concurrent access to configuration
func TestConcurrentConfigAccess(t *testing.T) {
	cfg := config.LoadDefaults()
	require.NotNil(t, cfg)
	
	const numGoroutines = 100
	const numIterations = 10
	
	done := make(chan bool, numGoroutines)
	
	// Launch concurrent goroutines accessing config
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()
			
			for j := 0; j < numIterations; j++ {
				// Test different component configs
				_, err := config.GetConfig(cfg, assets.ConfigParser{})
				if err != nil {
					t.Errorf("Failed to get assets config: %v", err)
					return
				}
				
				_, err = config.GetConfig(cfg, workspace.ConfigParser{})
				if err != nil {
					t.Errorf("Failed to get workspace config: %v", err)
					return
				}
			}
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(30 * time.Second):
			t.Fatal("Concurrent config access test timed out")
		}
	}
}