package schemas

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStableSchemasAreEmbeddedAndVersioned(t *testing.T) {
	for _, test := range []struct {
		kind    Kind
		version int
	}{
		{kind: Config, version: ConfigVersion},
		{kind: Data, version: DataVersion},
		{kind: Scenario, version: ScenarioVersion},
	} {
		data, err := Read(test.kind)
		if err != nil {
			t.Fatal(err)
		}
		var document map[string]any
		if err := json.Unmarshal(data, &document); err != nil {
			t.Fatalf("%s schema is not JSON: %v", test.kind, err)
		}
		properties, ok := document["properties"].(map[string]any)
		if !ok {
			t.Fatalf("%s schema has no properties", test.kind)
		}
		version, ok := properties["schema_version"].(map[string]any)["const"].(float64)
		if !ok || int(version) != test.version {
			t.Fatalf("%s schema version = %v, want %d", test.kind, version, test.version)
		}
		if id, _ := document["$id"].(string); !strings.Contains(id, "v1") {
			t.Fatalf("%s schema has unstable id %q", test.kind, id)
		}
	}
}
