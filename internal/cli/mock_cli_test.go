package cli

import (
	"errors"
	"github.com/spf13/cobra"
)

// MockRootCommand is a mock implementation of RootCommand for testing
type MockRootCommand struct {
	ExecuteError error
	RootCmd      *cobra.Command
}

// Execute mocks the Execute method of RootCommand
func (m *MockRootCommand) Execute() error {
	return m.ExecuteError
}

// MockInitializeApp allows tests to control the behavior of InitializeApp
var MockInitializeApp func() (*RootCommand, error)

// Original function reference
var originalInitializeApp = InitializeApp

// SetupMockInitializeApp sets up mocking for InitializeApp
func SetupMockInitializeApp() {
	// Save original function
	originalInitializeApp = InitializeApp

	// Replace with mock function
	InitializeApp = func() (*RootCommand, error) {
		if MockInitializeApp != nil {
			return MockInitializeApp()
		}
		return nil, errors.New("MockInitializeApp not set")
	}
}

// ResetMockInitializeApp restores the original InitializeApp function
func ResetMockInitializeApp() {
	InitializeApp = originalInitializeApp
}
