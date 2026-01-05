package pos

import (
	"testing"
	"time"
)

func TestDumpAndRestore(t *testing.T) {
	workspaceID := "test_workspace"
	// Use a fixed "now" time for testing
	// minTimeOffset is -18h, so minTime will be now - 18h
	// Test data times should be after minTime to be returned as-is
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		data map[string]time.Time
	}{
		{
			name: "empty position",
			data: map[string]time.Time{},
		},
		{
			name: "single entry",
			data: map[string]time.Time{
				"key1": now.Add(-1 * time.Hour),
			},
		},
		{
			name: "multiple entries",
			data: map[string]time.Time{
				"key1": now.Add(-1 * time.Hour),
				"key2": now.Add(-2 * time.Hour),
				"key3": now.Add(-3 * time.Hour),
			},
		},
		{
			name: "with nanoseconds",
			data: map[string]time.Time{
				"nano": now.Add(-1*time.Hour + 123456789*time.Nanosecond),
			},
		},
		{
			name: "different timezones",
			data: map[string]time.Time{
				"utc": now.Add(-1 * time.Hour).UTC(),
				"jst": now.Add(-2 * time.Hour).In(time.FixedZone("JST", 9*60*60)),
			},
		},
		{
			name: "special characters in keys",
			data: map[string]time.Time{
				"key/with/slashes": now.Add(-1 * time.Hour),
				"key-with-dashes":  now.Add(-2 * time.Hour),
				"key_with_under":   now.Add(-3 * time.Hour),
				"key.with.dots":    now.Add(-4 * time.Hour),
				"key:with:colons":  now.Add(-5 * time.Hour),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Pos using At with a fixed minTime
			original := At(workspaceID, now.Add(minTimeOffset))
			for k, v := range tt.data {
				original.Store(k, v)
			}

			// Dump the position
			dumped, err := original.Dump()
			if err != nil {
				t.Fatalf("Failed to dump position: %v", err)
			}

			// Restore from the dumped data using RestoreAt
			restored, err := RestoreAt(workspaceID, dumped, now)
			if err != nil {
				t.Fatalf("Failed to restore position: %v", err)
			}

			// Verify that all keys and values match
			for k, expectedTime := range tt.data {
				restoredTime := restored.Load(k)
				if !restoredTime.Equal(expectedTime) {
					t.Errorf("Time mismatch for key %q: got %v, want %v", k, restoredTime, expectedTime)
				}
			}

			// Verify that loading non-existent keys returns minTime
			nonExistentKey := "non_existent_key_12345"
			defaultTime := restored.Load(nonExistentKey)
			expectedMinTime := now.Add(minTimeOffset)
			if !defaultTime.Equal(expectedMinTime) {
				t.Errorf("Loading non-existent key should return minTime %v, got %v", expectedMinTime, defaultTime)
			}
		})
	}
}

func TestRestore_InvalidJSON(t *testing.T) {
	workspaceID := "test_workspace"
	invalidJSON := []byte(`{"invalid": "not a time"}`)
	_, err := Restore(workspaceID, invalidJSON)
	if err == nil {
		t.Error("Expected error when restoring invalid JSON, got nil")
	}
}

func TestRestore_EmptyBytes(t *testing.T) {
	workspaceID := "test_workspace"
	emptyJSON := []byte(`{}`)
	pos, err := Restore(workspaceID, emptyJSON)
	if err != nil {
		t.Fatalf("Failed to restore empty JSON: %v", err)
	}

	// Should return defaultTime for any key
	testKey := "test_key"
	loadedTime := pos.Load(testKey)
	if loadedTime.IsZero() {
		t.Error("Loading from empty position should return defaultTime, got zero time")
	}
}
