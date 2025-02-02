package clitooling

import (
	"github.com/spf13/cobra"
)

// RunParentPersistentPreRun always runs parent's PersistentPreRun.
func RunParentPersistentPreRun(cmd *cobra.Command, args []string) error {
	if p := cmd.Parent(); p != nil {
		return p.PersistentPreRunE(p, args)
	}
	return nil
}
