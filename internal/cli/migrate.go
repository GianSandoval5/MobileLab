package cli

import (
	"fmt"
	"path/filepath"

	"github.com/mobilelab-dev/mobilelab/internal/migration"
)

func (r Runner) migrateCommand(args []string) error {
	check := false
	if len(args) == 1 && args[0] == "--check" {
		check = true
	} else if len(args) != 0 {
		return fmt.Errorf("usage: mobilelab migrate [--check]")
	}
	plan, err := migration.Build(r.Dir)
	if err != nil {
		return fmt.Errorf("unable to prepare project migration: %w", err)
	}
	changes := plan.ChangeCount()
	if changes == 0 {
		fmt.Fprintln(r.Out, "Project schemas are current.")
		return nil
	}
	for _, document := range plan.Documents {
		if document.NeedsMigration {
			relative, relErr := filepath.Rel(r.Dir, document.Path)
			if relErr != nil {
				relative = document.Path
			}
			fmt.Fprintf(r.Out, "↑ %s: %s v%d → v%d\n", relative, document.Kind, document.FromVersion, document.ToVersion)
		}
	}
	if check {
		return fmt.Errorf("%d document(s) require migration; run 'mobilelab migrate'", changes)
	}
	if err := plan.Apply(); err != nil {
		return err
	}
	fmt.Fprintf(r.Out, "Migrated %d document(s).\n", changes)
	return nil
}
