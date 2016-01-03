package yaml

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestCreate_reparseYAML(t *testing.T) {
	cfg, err := ParseString(``)
	if err != nil {
		t.Fatal(err)
	}

	yamlBytes, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Parse(yamlBytes); err != nil {
		t.Fatal(err)
	}
}

func TestYAML_UnmarshalMapEqualSlice(t *testing.T) {
	v := NewMapEqualSlice(map[string]string{"a": "a", "b": "b"})
	yamlBytes, err := yaml.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	want := "a: a\nb: b\n"
	if string(yamlBytes) != want {
		t.Errorf("got %q, want %q", yamlBytes, want)
	}

	var v2 MapEqualSlice
	if err := yaml.Unmarshal(yamlBytes, &v2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v2, v) {
		t.Errorf("got %+v, want %+v", v2, v)
	}
}

func TestYAML_UnmarshalCommand(t *testing.T) {
	v := NewCommand("a", "b")
	yamlBytes, err := yaml.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	want := "- a\n- b\n"
	if string(yamlBytes) != want {
		t.Errorf("got %q, want %q", yamlBytes, want)
	}

	var v2 Command
	if err := yaml.Unmarshal(yamlBytes, &v2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v2, v) {
		t.Errorf("got %+v, want %+v", v2, v)
	}
}

func TestYAML_UnmarshalPluginslice(t *testing.T) {
	v := Pluginslice{}.
		WithAppendedPlugin("k1", Plugin{Container: Container{Image: "a"}}).
		WithAppendedPlugin("k2", Plugin{Container: Container{Image: "b"}})
	yamlBytes, err := yaml.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	wantSubstrings := []string{"k1:\n  image: a\n", "\nk2:\n  image: b\n"}
	for _, want := range wantSubstrings {
		if !strings.Contains(string(yamlBytes), want) {
			t.Errorf("got %q, want substring %q", yamlBytes, want)
		}
	}

	var v2 Pluginslice
	if err := yaml.Unmarshal(yamlBytes, &v2); err != nil {
		t.Fatal(err)
	}

	// Compare as JSON to treat []string{} as []string(nil).
	j1, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	j2, err := json.Marshal(v2)
	if err != nil {
		t.Fatal(err)
	}
	if string(j1) != string(j2) {
		t.Errorf("got %s, want %s", j1, j2)
	}
}

func TestYAML_UnmarshalBuildStep(t *testing.T) {
	v := BuildStep{}.
		WithAppendedBuild("k1", Build{Container: Container{Image: "a"}}).
		WithAppendedBuild("k2", Build{Container: Container{Image: "b"}})
	yamlBytes, err := yaml.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	wantSubstrings := []string{"k1:\n  image: a\n", "\nk2:\n  image: b\n"}
	for _, want := range wantSubstrings {
		if !strings.Contains(string(yamlBytes), want) {
			t.Errorf("got %q, want substring %q", yamlBytes, want)
		}
	}

	var v2 BuildStep
	if err := yaml.Unmarshal(yamlBytes, &v2); err != nil {
		t.Fatal(err)
	}

	// Compare as JSON to treat []string{} as []string(nil).
	j1, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	j2, err := json.Marshal(v2)
	if err != nil {
		t.Fatal(err)
	}
	if string(j1) != string(j2) {
		t.Errorf("got %s, want %s", j1, j2)
	}
}

func TestBuildStep_MarshalYAML(t *testing.T) {
	build := Build{Container: Container{Image: "a"}}

	yamlBytes, err := yaml.Marshal(BuildStep{parts: []Build{build}})
	if err != nil {
		t.Fatal(err)
	}

	want, err := yaml.Marshal(build)
	if err != nil {
		t.Fatal(err)
	}

	if string(yamlBytes) != string(want) {
		t.Errorf("got %q, want %q", yamlBytes, want)
	}
}
