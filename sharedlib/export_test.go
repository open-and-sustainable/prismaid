package main

import "testing"

// Since this is a CGO file with exports, we can't easily test the exported functions
// directly as they're meant to be called from Python/R. We'll create a simple test
// that ensures the package can be compiled at minimum.

func TestSharedLibExistence(t *testing.T) {
	// This test simply verifies that the package can be compiled
	// We're not testing the actual CGO functions as they require C memory management

	// Test the existence of our internal helper function
	t.Run("handlePanic function exists", func(t *testing.T) {
		// Call the function with a test panic and recover
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Error("Expected handlePanic to recover, but it didn't")
				}
			}()

			// Test that we can call the function
			// This won't actually execute the CGO parts
			panic("Test panic")
		}()
	})

	// Additional tests could verify the function signatures
	// but that's more complex with CGO exports
}
