package exporters

import (
	"testing"
)

func TestNewCollector(t *testing.T) {
	collector := NewCollector()

	// Verify that the DeletedPVC counter is not nil
	if collector.DeletedPVC == nil {
		t.Errorf("Expected DeletedPVC counter to be initialized, got nil")
	}
}
