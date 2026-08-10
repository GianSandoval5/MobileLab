package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mobilelab-dev/mobilelab/internal/config"
	"github.com/mobilelab-dev/mobilelab/internal/datastore"
)

func (r Runner) databaseCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: mobilelab db init | seed | reset | status [--json]")
	}
	switch args[0] {
	case "init":
		if len(args) != 1 {
			return errors.New("usage: mobilelab db init")
		}
		return r.databaseInit(ctx)
	case "seed", "reset":
		if len(args) != 1 {
			return fmt.Errorf("usage: mobilelab db %s", args[0])
		}
		mode := datastore.SeedUpsert
		if args[0] == "reset" {
			mode = datastore.SeedReset
		}
		cfg, store, err := r.openBusinessDatabase()
		if err != nil {
			return err
		}
		defer store.Close()
		if err := store.Seed(ctx, cfg, filepath.Join(r.Dir, "mobilelab"), mode); err != nil {
			return fmt.Errorf("unable to %s business data: %w", args[0], err)
		}
		verb := "Seeded"
		if mode == datastore.SeedReset {
			verb = "Reset and seeded"
		}
		fmt.Fprintf(r.Out, "%s mobilelab/%s.\n", verb, datastore.DatabaseFilename)
		return r.printDatabaseStatus(ctx, cfg, store, false)
	case "status":
		if len(args) > 2 || (len(args) == 2 && args[1] != "--json") {
			return errors.New("usage: mobilelab db status [--json]")
		}
		cfg, store, err := r.openBusinessDatabase()
		if err != nil {
			return err
		}
		defer store.Close()
		return r.printDatabaseStatus(ctx, cfg, store, len(args) == 2)
	default:
		return fmt.Errorf("unknown db command %q; usage: mobilelab db init | seed | reset | status [--json]", args[0])
	}
}

func (r Runner) databaseInit(ctx context.Context) error {
	coreConfig, err := config.Load(r.configPath())
	if err != nil {
		return fmt.Errorf("unable to initialize business data: %w", err)
	}
	configPath := datastore.ConfigPath(r.Dir)
	created := false
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		dataYAML := "schema_version: 1\nresources:\n  items:\n    path: /api/items\n    id: id\n    seed: seeds/items.json\n"
		starter, parseErr := datastore.Parse([]byte(dataYAML))
		if parseErr != nil {
			return fmt.Errorf("prepare starter data configuration: %w", parseErr)
		}
		if conflictErr := starter.ValidateEndpoints(coreConfig.Endpoints); conflictErr != nil {
			return fmt.Errorf("starter /api/items route conflicts with mobilelab.yaml: %w; create mobilelab/data.yaml with another path", conflictErr)
		}
		seedDirectory := filepath.Join(r.Dir, "mobilelab", "seeds")
		if err := os.MkdirAll(seedDirectory, 0o755); err != nil {
			return fmt.Errorf("create data seed directory: %w", err)
		}
		if err := writeExclusive(filepath.Join(seedDirectory, "items.json"), []byte("[]\n")); err != nil {
			return err
		}
		if err := writeExclusive(configPath, []byte(dataYAML)); err != nil {
			return err
		}
		created = true
	} else if err != nil {
		return fmt.Errorf("inspect data configuration: %w", err)
	}
	cfg, store, err := r.openBusinessDatabase()
	if err != nil {
		return err
	}
	defer store.Close()
	if err := store.Seed(ctx, cfg, filepath.Join(r.Dir, "mobilelab"), datastore.SeedEmpty); err != nil {
		return fmt.Errorf("initialize business data: %w", err)
	}
	if created {
		fmt.Fprintln(r.Out, "Created mobilelab/data.yaml and mobilelab/seeds/items.json.")
	}
	fmt.Fprintf(r.Out, "Business database ready at mobilelab/%s.\n", datastore.DatabaseFilename)
	return r.printDatabaseStatus(ctx, cfg, store, false)
}

func (r Runner) openBusinessDatabase() (datastore.Config, *datastore.Store, error) {
	coreConfig, err := config.Load(r.configPath())
	if err != nil {
		return datastore.Config{}, nil, fmt.Errorf("load MobileLab project: %w", err)
	}
	cfg, configured, err := datastore.LoadOptional(r.Dir)
	if err != nil {
		return datastore.Config{}, nil, err
	}
	if !configured {
		return datastore.Config{}, nil, errors.New("mobilelab/data.yaml does not exist; run 'mobilelab db init'")
	}
	if err := cfg.ValidateEndpoints(coreConfig.Endpoints); err != nil {
		return datastore.Config{}, nil, err
	}
	store, err := datastore.Open(datastore.DatabasePath(r.Dir))
	if err != nil {
		return datastore.Config{}, nil, err
	}
	return cfg, store, nil
}

func (r Runner) printDatabaseStatus(ctx context.Context, cfg datastore.Config, store *datastore.Store, asJSON bool) error {
	counts, err := store.Counts(ctx, cfg)
	if err != nil {
		return err
	}
	if asJSON {
		payload := map[string]any{"database": filepath.Join("mobilelab", datastore.DatabaseFilename), "schema_version": cfg.SchemaVersion, "resources": counts}
		encoded, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintf(r.Out, "%s\n", encoded)
		return nil
	}
	fmt.Fprintf(r.Out, "Business data resources (%s):\n", filepath.Join("mobilelab", datastore.DatabaseFilename))
	for _, name := range cfg.Names() {
		resource := cfg.Resources[name]
		kind := "collection"
		if resource.Singleton {
			kind = "singleton"
		}
		fmt.Fprintf(r.Out, "✓ %-18s %4d document(s)  %s  %s\n", name, counts[name], strings.ToUpper(kind), resource.Path)
	}
	return nil
}

func writeExclusive(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}
	return nil
}
