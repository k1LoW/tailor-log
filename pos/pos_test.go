package pos

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDumpAndRestore(t *testing.T) {
	workspaceID := "test_workspace"
	// Use a fixed minTime that is before all test times
	minTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
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
				"key1": time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "multiple entries",
			data: map[string]time.Time{
				"key1": time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				"key2": time.Date(2024, 2, 15, 9, 30, 0, 0, time.UTC),
				"key3": time.Date(2024, 3, 20, 18, 45, 30, 0, time.UTC),
			},
		},
		{
			name: "with nanoseconds",
			data: map[string]time.Time{
				"nano": time.Date(2024, 6, 15, 10, 20, 30, 123456789, time.UTC),
			},
		},
		{
			name: "different timezones",
			data: map[string]time.Time{
				"utc": time.Date(2024, 5, 10, 14, 0, 0, 0, time.UTC),
				"jst": time.Date(2024, 5, 10, 23, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
			},
		},
		{
			name: "special characters in keys",
			data: map[string]time.Time{
				"key/with/slashes": time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
				"key-with-dashes":  time.Date(2024, 7, 2, 0, 0, 0, 0, time.UTC),
				"key_with_under":   time.Date(2024, 7, 3, 0, 0, 0, 0, time.UTC),
				"key.with.dots":    time.Date(2024, 7, 4, 0, 0, 0, 0, time.UTC),
				"key:with:colons":  time.Date(2024, 7, 5, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Pos using From with a fixed minTime
			original := From(workspaceID, minTime)
			for k, v := range tt.data {
				original.Store(k, v)
			}

			// Dump the position
			dumped, err := original.Dump()
			if err != nil {
				t.Fatalf("Failed to dump position: %v", err)
			}

			// Restore from the dumped data using From
			restored := From(workspaceID, minTime)
			var m map[string]time.Time
			if err := json.Unmarshal(dumped, &m); err != nil {
				t.Fatalf("Failed to unmarshal dumped data: %v", err)
			}
			for k, v := range m {
				restored.Store(k, v)
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
			if !defaultTime.Equal(minTime) {
				t.Errorf("Loading non-existent key should return minTime, got %v", defaultTime)
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
