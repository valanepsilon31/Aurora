package main

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

// TestCommandFlagsIsolation ensures each command only accesses its own flags
// This catches issues like the backup command trying to access config's "reset" flag
func TestCommandFlagsIsolation(t *testing.T) {
	tests := []struct {
		name     string
		cmd      *cobra.Command
		flags    []string
		badFlags []string
	}{
		{
			name:     "backup command flags",
			cmd:      backupCmd,
			flags:    []string{"validate", "thread"},
			badFlags: []string{"reset"}, // belongs to config command
		},
		{
			name:     "config command flags",
			cmd:      configCmd,
			flags:    []string{"reset"},
			badFlags: []string{"validate", "thread"}, // belongs to backup command
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify expected flags exist and are accessible
			for _, flag := range tt.flags {
				f := tt.cmd.Flags().Lookup(flag)
				if f == nil {
					t.Errorf("expected flag %q to be defined on %s", flag, tt.cmd.Name())
				}
			}

			// Verify flags from other commands are NOT accessible
			for _, flag := range tt.badFlags {
				f := tt.cmd.Flags().Lookup(flag)
				if f != nil {
					t.Errorf("flag %q should NOT be defined on %s (belongs to another command)", flag, tt.cmd.Name())
				}
			}
		})
	}
}

// TestCommandsCanExecuteWithHelp verifies commands don't panic when showing help
// This is a basic smoke test for command initialization
func TestCommandsCanExecuteWithHelp(t *testing.T) {
	commands := []*cobra.Command{
		configCmd,
		backupCmd,
		penumbraCmd,
	}

	for _, cmd := range commands {
		t.Run(cmd.Name()+" --help", func(t *testing.T) {
			// Reset command output
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs([]string{"--help"})

			// Should not panic or error on help
			err := cmd.Execute()
			if err != nil {
				t.Errorf("command %s --help failed: %v", cmd.Name(), err)
			}
		})
	}
}

// TestBackupFlagAccess specifically tests the flag access that caused the original bug
// The bug was: backup command calling code that tried to access "reset" flag
func TestBackupFlagAccess(t *testing.T) {
	// Simulate what happens when backup command runs and tries to access its flags
	cmd := backupCmd

	// These should work without error
	_, err := cmd.Flags().GetBool("validate")
	if err != nil {
		t.Errorf("backup command should be able to access 'validate' flag: %v", err)
	}

	_, err = cmd.Flags().GetInt("thread")
	if err != nil {
		t.Errorf("backup command should be able to access 'thread' flag: %v", err)
	}

	// This should fail - reset belongs to config command
	_, err = cmd.Flags().GetBool("reset")
	if err == nil {
		t.Error("backup command should NOT be able to access 'reset' flag (belongs to config)")
	}
}

// TestConfigFlagAccess tests config command flag access
func TestConfigFlagAccess(t *testing.T) {
	cmd := configCmd

	_, err := cmd.Flags().GetBool("reset")
	if err != nil {
		t.Errorf("config command should be able to access 'reset' flag: %v", err)
	}

	// These should fail - belongs to backup command
	_, err = cmd.Flags().GetBool("validate")
	if err == nil {
		t.Error("config command should NOT be able to access 'validate' flag (belongs to backup)")
	}
}
