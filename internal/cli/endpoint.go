package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mobilelab-dev/mobilelab/internal/config"
	"github.com/mobilelab-dev/mobilelab/internal/endpoint"
)

func (r Runner) endpointCommand(ctx context.Context, args []string) error {
	selection, options, err := parseDeviceFlags(args, false, true)
	if err != nil {
		return fmt.Errorf("usage: mobilelab endpoint [--platform android|ios] [--device id] [--json]: %w", err)
	}
	configuration, err := config.Load(r.configPath())
	if err != nil {
		return err
	}
	var result endpoint.Result
	if selection.Platform == "" && selection.ID == "" {
		result, err = endpoint.Resolve(configuration, "", nil)
	} else {
		_, detected, selectErr := r.selectDevice(ctx, selection, "")
		if selectErr != nil {
			return selectErr
		}
		result, err = endpoint.Resolve(configuration, selection.Platform, &detected)
	}
	if err != nil {
		return fmt.Errorf("unable to resolve MobileLab endpoint: %w", err)
	}
	if options.JSON {
		encoder := json.NewEncoder(r.Out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}
	fmt.Fprintln(r.Out, result.URL)
	if result.Note != "" {
		fmt.Fprintf(r.Err, "Note: %s\n", result.Note)
	}
	return nil
}
