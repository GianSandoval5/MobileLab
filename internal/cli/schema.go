package cli

import (
	"fmt"

	"github.com/mobilelab-dev/mobilelab/schemas"
)

func (r Runner) schemaCommand(args []string) error {
	if len(args) != 1 || (args[0] != string(schemas.Config) && args[0] != string(schemas.Data) && args[0] != string(schemas.Scenario)) {
		return fmt.Errorf("usage: mobilelab schema config | data | scenario")
	}
	data, err := schemas.Read(schemas.Kind(args[0]))
	if err != nil {
		return err
	}
	if _, err := r.Out.Write(data); err != nil {
		return fmt.Errorf("write %s schema: %w", args[0], err)
	}
	return nil
}
