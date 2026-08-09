package doctor

import "testing"

func TestNodeMajor(t *testing.T) {
	for input, want := range map[string]int{"v22.23.2": 22, "18.0.0": 18} {
		got, err := nodeMajor(input)
		if err != nil || got != want {
			t.Fatalf("nodeMajor(%q) = %d, %v; want %d", input, got, err, want)
		}
	}
	if _, err := nodeMajor("unknown"); err == nil {
		t.Fatal("expected invalid Node.js version")
	}
}
