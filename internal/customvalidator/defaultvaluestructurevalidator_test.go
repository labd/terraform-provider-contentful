package customvalidator

import (
	"testing"
)

func TestDefaultValueStructureValidator_Creation(t *testing.T) {
	// Basic test to ensure the validator can be created
	validator := DefaultValueStructure()
	if validator == nil {
		t.Error("DefaultValueStructure() should not return nil")
	}
}
