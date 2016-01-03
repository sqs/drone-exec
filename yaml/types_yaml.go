package yaml

import (
	"fmt"
	"sort"
	"strings"

	"github.com/flynn/go-shlex"
	"gopkg.in/yaml.v2"
)

func NewCommand(parts ...string) Command { return Command{parts} }

type Command struct {
	parts []string
}

func (s Command) MarshalYAML() (interface{}, error) {
	return s.parts, nil
}

func (s *Command) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stringType string
	err := unmarshal(&stringType)
	if err == nil {
		s.parts, err = shlex.Split(stringType)
		return err
	}

	var sliceType []string
	err = unmarshal(&sliceType)
	if err == nil {
		s.parts = sliceType
		return nil
	}

	return err
}

func (s *Command) Slice() []string {
	return s.parts
}

type MapEqualSlice struct {
	parts []string
}

func NewMapEqualSlice(m map[string]string) MapEqualSlice {
	o := MapEqualSlice{parts: make([]string, 0, len(m))}
	for k, v := range m {
		o.parts = append(o.parts, k+"="+v)
	}
	sort.Strings(o.parts)
	return o
}

func (s MapEqualSlice) MarshalYAML() (interface{}, error) {
	m := make(map[string]string, len(s.parts))
	for _, part := range s.parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 1 {
			kv = append(kv, "")
		}
		m[kv[0]] = kv[1]
	}
	return m, nil
}

func (s *MapEqualSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal(&s.parts)
	if err == nil {
		return nil
	}

	var mapType map[string]string

	err = unmarshal(&mapType)
	if err != nil {
		return err
	}

	for k, v := range mapType {
		s.parts = append(s.parts, strings.Join([]string{k, v}, "="))
	}
	sort.Strings(s.parts)

	return nil
}

func (s *MapEqualSlice) Slice() []string {
	return s.parts
}

// Stringorslice represents a string or an array of strings.
// TODO use docker/docker/pkg/stringutils.StrSlice once 1.9.x is released.
type Stringorslice struct {
	parts []string
}

// MarshalYAML implements the Marshaller interface.
func (s Stringorslice) MarshalYAML() (interface{}, error) {
	return s.parts, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *Stringorslice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var sliceType []string
	err := unmarshal(&sliceType)
	if err == nil {
		s.parts = sliceType
		return nil
	}

	var stringType string
	err = unmarshal(&stringType)
	if err == nil {
		sliceType = make([]string, 0, 1)
		s.parts = append(sliceType, string(stringType))
		return nil
	}
	return err
}

// Len returns the number of parts of the Stringorslice.
func (s *Stringorslice) Len() int {
	if s == nil {
		return 0
	}
	return len(s.parts)
}

// Slice gets the parts of the StrSlice as a Slice of string.
func (s *Stringorslice) Slice() []string {
	if s == nil {
		return nil
	}
	return s.parts
}

// Pluginslice is a slice of Plugins with a custom Yaml
// unarmshal function to preserve ordering.
type Pluginslice struct {
	parts []Plugin
	keys  []string
}

func (s Pluginslice) MarshalYAML() (interface{}, error) {
	obj := yaml.MapSlice{}
	for i, p := range s.parts {
		obj = append(obj, yaml.MapItem{Key: s.keys[i], Value: p})
	}
	return obj, nil
}

func (s *Pluginslice) UnmarshalYAML(unmarshal func(interface{}) error) error {

	// unmarshal the yaml into the generic
	// mapSlice type to preserve ordering.
	obj := yaml.MapSlice{}
	err := unmarshal(&obj)
	if err != nil {
		return err
	}

	// unarmshals each item in the mapSlice,
	// unmarshal and append to the slice.
	err = unmarshalYaml(obj, func(key string, val []byte) error {
		plugin := Plugin{}
		err := yaml.Unmarshal(val, &plugin)
		if err != nil {
			return err
		}
		if len(plugin.Image) == 0 {
			plugin.Image = key
		}
		s.parts = append(s.parts, plugin)
		s.keys = append(s.keys, key)
		return nil
	})
	return err
}

func (s *Pluginslice) Slice() []Plugin {
	return s.parts
}

// WithAppendedPlugin copies s and adds a new plugin with the given
// key.
func (s Pluginslice) WithAppendedPlugin(key string, p Plugin) Pluginslice {
	return Pluginslice{append(s.parts, p), append(s.keys, key)}
}

// ContainerSlice is a slice of Containers with a custom
// Yaml unarmshal function to preserve ordering.
type Containerslice struct {
	parts []Container
	keys  []string
}

func (s Containerslice) MarshalYAML() (interface{}, error) {
	obj := yaml.MapSlice{}
	for i, p := range s.parts {
		obj = append(obj, yaml.MapItem{Key: s.keys[i], Value: p})
	}
	return obj, nil
}

func (s *Containerslice) UnmarshalYAML(unmarshal func(interface{}) error) error {

	// unmarshal the yaml into the generic
	// mapSlice type to preserve ordering.
	obj := yaml.MapSlice{}
	err := unmarshal(&obj)
	if err != nil {
		return err
	}

	// unarmshals each item in the mapSlice,
	// unmarshal and append to the slice.
	return unmarshalYaml(obj, func(key string, val []byte) error {
		ctr := Container{}
		err := yaml.Unmarshal(val, &ctr)
		if err != nil {
			return err
		}
		if len(ctr.Image) == 0 {
			ctr.Image = key
		}
		s.parts = append(s.parts, ctr)
		s.keys = append(s.keys, key)
		return nil
	})
}

func (s *Containerslice) Slice() []Container {
	return s.parts
}

// WithAppendedContainer copies s and adds a new container with the
// given key.
func (s Containerslice) WithAppendedContainer(key string, ctr Container) Containerslice {
	return Containerslice{append(s.parts, ctr), append(s.keys, key)}
}

// BuildStep holds the build step configuration using a custom
// Yaml unarmshal function to preserve ordering.
type BuildStep struct {
	parts []Build
	keys  []string
}

func (s BuildStep) MarshalYAML() (interface{}, error) {
	if s.parts != nil && s.keys == nil {
		return s.parts[0], nil
	}
	obj := yaml.MapSlice{}
	for i, p := range s.parts {
		obj = append(obj, yaml.MapItem{Key: s.keys[i], Value: p})
	}
	return obj, nil
}

func (s *BuildStep) UnmarshalYAML(unmarshal func(interface{}) error) error {
	build := Build{}
	err := unmarshal(&build)
	if err != nil {
		return err
	}
	if build.Image != "" {
		s.parts = append(s.parts, build)
		return nil
	}

	// unmarshal the yaml into the generic
	// mapSlice type to preserve ordering.
	obj := yaml.MapSlice{}
	if err := unmarshal(&obj); err != nil {
		return err
	}

	// unarmshals each item in the mapSlice,
	// unmarshal and append to the slice.
	return unmarshalYaml(obj, func(key string, val []byte) error {
		build := Build{}
		err := yaml.Unmarshal(val, &build)
		if err != nil {
			return err
		}
		s.parts = append(s.parts, build)
		s.keys = append(s.keys, key)
		return nil
	})
}

func (s *BuildStep) Slice() []Build {
	return s.parts
}

// WithAppendedBuild copies s and adds a new Build with the
// given key.
func (s BuildStep) WithAppendedBuild(key string, build Build) BuildStep {
	if s.parts != nil && s.keys == nil {
		panic("can't append build to a non-multi-build section")
	}
	return BuildStep{append(s.parts, build), append(s.keys, key)}
}

// emitter defines the callback function used for
// generic yaml parsing. It emits back a raw byte
// slice for custom unmarshalling into a structure.
type unmarshal func(string, []byte) error

// unmarshalYaml is a helper function that removes
// some of the boilerplate from unmarshalling
// complex map slices.
func unmarshalYaml(v yaml.MapSlice, emit unmarshal) error {
	for _, vv := range v {
		// re-marshal the interface{} back to
		// a raw yaml value
		val, err := yaml.Marshal(&vv.Value)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%v", vv.Key)

		// unmarshal the raw value using the
		// callback function.
		if err := emit(key, val); err != nil {
			return err
		}
	}
	return nil
}
