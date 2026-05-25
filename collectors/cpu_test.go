package collectors

import (
	"math"
	"testing"
)

func TestRoundTo1(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0.0, 0.0},
		{1.0, 1.0},
		{1.55, 1.6},
		{1.54, 1.5},
		{99.95, 100.0},
		{33.33, 33.3},
		{0.05, 0.1},
		{0.04, 0.0},
		{100.0, 100.0},
	}

	for _, tc := range tests {
		got := roundTo1(tc.input)
		if math.Abs(got-tc.expected) > 0.0001 {
			t.Errorf("roundTo1(%v) = %v, want %v", tc.input, got, tc.expected)
		}
	}
}

func TestRoundTo2(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0.0, 0.0},
		{1.0, 1.0},
		{1.555, 1.56},
		{1.554, 1.55},
		{0.005, 0.01},
		{0.004, 0.0},
		{15.937, 15.94},
		{1024.0, 1024.0},
	}

	for _, tc := range tests {
		got := roundTo2(tc.input)
		if math.Abs(got-tc.expected) > 0.0001 {
			t.Errorf("roundTo2(%v) = %v, want %v", tc.input, got, tc.expected)
		}
	}
}

func TestRoundSlice(t *testing.T) {
	input := []float64{1.55, 2.34, 0.05, 99.95}
	expected := []float64{1.6, 2.3, 0.1, 100.0}

	got := roundSlice(input)

	if len(got) != len(expected) {
		t.Fatalf("roundSlice length = %d, want %d", len(got), len(expected))
	}
	for i, v := range got {
		if math.Abs(v-expected[i]) > 0.0001 {
			t.Errorf("roundSlice[%d] = %v, want %v", i, v, expected[i])
		}
	}
}

func TestRoundSliceEmpty(t *testing.T) {
	got := roundSlice(nil)
	if len(got) != 0 {
		t.Errorf("roundSlice(nil) = %v, want empty slice", got)
	}
}

func TestRoundSliceDoesNotModifyOriginal(t *testing.T) {
	input := []float64{1.55, 2.34}
	original := make([]float64, len(input))
	copy(original, input)

	_ = roundSlice(input)

	for i, v := range input {
		if v != original[i] {
			t.Errorf("roundSlice modified original at index %d: got %v, want %v", i, v, original[i])
		}
	}
}

func TestCollectCPU(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	stats, err := CollectCPU()
	if err != nil {
		t.Fatalf("CollectCPU() error: %v", err)
	}
	if stats == nil {
		t.Fatal("CollectCPU() returned nil stats")
	}

	if stats.CoreCount <= 0 {
		t.Errorf("CoreCount = %d, want > 0", stats.CoreCount)
	}
	if stats.UsagePercent < 0 || stats.UsagePercent > 100 {
		t.Errorf("UsagePercent = %v, want [0, 100]", stats.UsagePercent)
	}
	if stats.LoadAvg1min < 0 {
		t.Errorf("LoadAvg1min = %v, want >= 0", stats.LoadAvg1min)
	}
	if len(stats.PerCorePercent) != stats.CoreCount {
		t.Errorf("PerCorePercent length = %d, want %d", len(stats.PerCorePercent), stats.CoreCount)
	}
	for i, p := range stats.PerCorePercent {
		if p < 0 || p > 100 {
			t.Errorf("PerCorePercent[%d] = %v, want [0, 100]", i, p)
		}
	}
}
