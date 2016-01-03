package yaml

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestParseString(t *testing.T) {
	v, err := ParseString(``)
	if err != nil {
		t.Fatal(err)
	}

	yamlBytes, err := yaml.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Parse(yamlBytes); err != nil {
		t.Fatal(err)
	}
}

func TestBuilds(t *testing.T) {
	tests := map[string]struct {
		orig      Builds
		wantYAML  string
		wantFinal *Builds // if nil, uses orig
	}{
		"single build with no key": {
			orig:     Builds{BuildItem{Key: "", Build: Build{Container: Container{Image: "img1"}}}},
			wantYAML: "image: img1\n",
		},
		"single build with key": {
			orig:     Builds{BuildItem{Key: "k1", Build: Build{Container: Container{Image: "img1"}}}},
			wantYAML: "k1:\n  image: img1\n",
		},
		"multiple builds": {
			orig: Builds{
				BuildItem{Key: "k1", Build: Build{Container: Container{Image: "img1"}}},
				BuildItem{Key: "k2", Build: Build{Container: Container{Image: "img2"}}},
			},
			wantYAML: "k1:\n  image: img1\nk2:\n  image: img2\n",
		},
	}
	for label, test := range tests {
		yamlBytes, err := yaml.Marshal(test.orig)
		if err != nil {
			t.Errorf("%s: Marshal: %s", label, err)
			continue
		}
		if string(yamlBytes) != test.wantYAML {
			t.Errorf("%s: got YAML %q, want %q", label, yamlBytes, test.wantYAML)
			continue
		}

		var final Builds
		if err := yaml.Unmarshal(yamlBytes, &final); err != nil {
			t.Errorf("%s: Unmarshal: %s", label, err)
			continue
		}
		if test.wantFinal == nil {
			test.wantFinal = &test.orig
		}
		if !reflect.DeepEqual(final, *test.wantFinal) {
			t.Errorf("%s: got %#v, want %#v", label, final, *test.wantFinal)
			continue
		}
	}
}

func TestCommand(t *testing.T) {
	tests := map[string]struct {
		origYAML      string
		want          Command
		wantFinalYAML string // if empty, uses origYAML
	}{
		"string command": {
			origYAML:      `a b c`,
			want:          Command{"a", "b", "c"},
			wantFinalYAML: "- a\n- b\n- c\n",
		},
		"array command": {
			origYAML: "- a\n- b\n- c\n",
			want:     Command{"a", "b", "c"},
		},
	}
	for label, test := range tests {
		var command Command
		if err := yaml.Unmarshal([]byte(test.origYAML), &command); err != nil {
			t.Errorf("%s: Unmarshal: %s", label, err)
			continue
		}
		if !reflect.DeepEqual(command, test.want) {
			t.Errorf("%s: got %#v, want %#v", label, command, test.want)
			continue
		}

		yamlBytes, err := yaml.Marshal(command)
		if err != nil {
			t.Errorf("%s: Marshal: %s", label, err)
			continue
		}
		if test.wantFinalYAML == "" {
			test.wantFinalYAML = test.origYAML
		}
		if string(yamlBytes) != test.wantFinalYAML {
			t.Errorf("%s: got YAML %q, want %q", label, yamlBytes, test.wantFinalYAML)
			continue
		}
	}
}

func TestContainers(t *testing.T) {
	tests := map[string]struct {
		orig      Containers
		wantYAML  string
		wantFinal *Containers // if nil, uses orig
	}{
		"single container with no image specified": {
			orig:      Containers{ContainerItem{Key: "k1", Container: Container{Image: ""}}},
			wantYAML:  "k1: {}\n",
			wantFinal: &Containers{ContainerItem{Key: "k1", Container: Container{Image: "k1"}}},
		},
		"single container with image specified": {
			orig:     Containers{ContainerItem{Key: "k1", Container: Container{Image: "img1"}}},
			wantYAML: "k1:\n  image: img1\n",
		},
		"multiple containers": {
			orig: Containers{
				ContainerItem{Key: "k1", Container: Container{Image: "img1"}},
				ContainerItem{Key: "k2", Container: Container{Image: "img2"}},
			},
			wantYAML: "k1:\n  image: img1\nk2:\n  image: img2\n",
		},
	}
	for label, test := range tests {
		yamlBytes, err := yaml.Marshal(test.orig)
		if err != nil {
			t.Errorf("%s: Marshal: %s", label, err)
			continue
		}
		if string(yamlBytes) != test.wantYAML {
			t.Errorf("%s: got YAML %q, want %q", label, yamlBytes, test.wantYAML)
			continue
		}

		var final Containers
		if err := yaml.Unmarshal(yamlBytes, &final); err != nil {
			t.Errorf("%s: Unmarshal: %s", label, err)
			continue
		}
		if test.wantFinal == nil {
			test.wantFinal = &test.orig
		}
		if !reflect.DeepEqual(final, *test.wantFinal) {
			t.Errorf("%s: got %#v, want %#v", label, final, *test.wantFinal)
			continue
		}
	}
}

func TestPlugins(t *testing.T) {
	tests := map[string]struct {
		orig      Plugins
		wantYAML  string
		wantFinal *Plugins // if nil, uses orig
	}{
		"single plugin with no image specified": {
			orig:      Plugins{PluginItem{Key: "k1", Plugin: Plugin{Container: Container{Image: ""}}}},
			wantYAML:  "k1: {}\n",
			wantFinal: &Plugins{PluginItem{Key: "k1", Plugin: Plugin{Container: Container{Image: "k1"}}}},
		},
		"single plugin with image specified": {
			orig:     Plugins{PluginItem{Key: "k1", Plugin: Plugin{Container: Container{Image: "img1"}}}},
			wantYAML: "k1:\n  image: img1\n",
		},
		"multiple plugins": {
			orig: Plugins{
				PluginItem{Key: "k1", Plugin: Plugin{Container: Container{Image: "img1"}}},
				PluginItem{Key: "k2", Plugin: Plugin{Container: Container{Image: "img2"}}},
			},
			wantYAML: "k1:\n  image: img1\nk2:\n  image: img2\n",
		},
	}
	for label, test := range tests {
		yamlBytes, err := yaml.Marshal(test.orig)
		if err != nil {
			t.Errorf("%s: Marshal: %s", label, err)
			continue
		}
		if string(yamlBytes) != test.wantYAML {
			t.Errorf("%s: got YAML %q, want %q", label, yamlBytes, test.wantYAML)
			continue
		}

		var final Plugins
		if err := yaml.Unmarshal(yamlBytes, &final); err != nil {
			t.Errorf("%s: Unmarshal: %s", label, err)
			continue
		}
		if test.wantFinal == nil {
			test.wantFinal = &test.orig
		}
		if !reflect.DeepEqual(final, *test.wantFinal) {
			t.Errorf("%s: got %#v, want %#v", label, final, *test.wantFinal)
			continue
		}
	}
}

func TestMapEqualSlice(t *testing.T) {
	tests := map[string]struct {
		origYAML      string
		want          MapEqualSlice
		wantFinalYAML string // if empty, uses origYAML
	}{
		"single": {
			origYAML: "k1: v1\n",
			want:     MapEqualSlice{"k1=v1"},
		},
		"multiple": {
			origYAML: "k1: v1\nk2: v2\nk3: v3\n",
			want:     MapEqualSlice{"k1=v1", "k2=v2", "k3=v3"},
		},
	}
	for label, test := range tests {
		var v MapEqualSlice
		if err := yaml.Unmarshal([]byte(test.origYAML), &v); err != nil {
			t.Errorf("%s: Unmarshal: %s", label, err)
			continue
		}
		if !reflect.DeepEqual(v, test.want) {
			t.Errorf("%s: got %#v, want %#v", label, v, test.want)
			continue
		}

		yamlBytes, err := yaml.Marshal(v)
		if err != nil {
			t.Errorf("%s: Marshal: %s", label, err)
			continue
		}
		if test.wantFinalYAML == "" {
			test.wantFinalYAML = test.origYAML
		}
		if string(yamlBytes) != test.wantFinalYAML {
			t.Errorf("%s: got YAML %q, want %q", label, yamlBytes, test.wantFinalYAML)
			continue
		}
	}
}
