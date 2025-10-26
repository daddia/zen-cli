package performance_test

import (
	"testing"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/assets"
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

// TestConfigLoadPerformance tests that config loading meets P95 ≤ 10ms requirement
func TestConfigLoadPerformance(t *testing.T) {
	const iterations = 1000
	const p95Threshold = 10 * time.Millisecond

	durations := make([]time.Duration, iterations)

	for i := 0; i < iterations; i++ {
		start := time.Now()
		cfg, err := config.Load()
		duration := time.Since(start)

		if err != nil {
			t.Fatal(err)
		}
		if cfg == nil {
			t.Fatal("config is nil")
		}

		durations[i] = duration
	}

	// Calculate P95
	// Simple approach: sort and take 95th percentile
	// For production, you'd use a proper percentile calculation
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

	// For this simple test, we'll use max as a proxy for P95
	// In a real implementation, you'd sort and calculate the actual P95
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

		if err != nil {
			t.Fatal(err)
		}
		if assetConfig.RepositoryURL == "" {
			t.Fatal("asset config is invalid")
		}

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

	// For this simple test, we'll use max as a proxy for P95
	if max > p95Threshold {
		t.Errorf("Component config parsing does not meet P95 ≤ %v requirement (max: %v)", p95Threshold, max)
	}
}
