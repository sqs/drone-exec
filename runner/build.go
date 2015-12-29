package runner

import (
	"errors"

	// log "github.com/Sirupsen/logrus"
	"github.com/drone/drone-exec/docker"
	"github.com/drone/drone-exec/parser"
	"github.com/drone/drone-exec/runner/script"
	"github.com/samalba/dockerclient"
)

var ErrNoImage = errors.New("Yaml must specify an image for every step")

// Default clone plugin.
var DefaultCloner = "plugins/drone-git"

// Default cache plugin.
var DefaultCacher = "plugins/drone-cache"

type Build struct {
	tree  *parser.Tree
	flags parser.NodeType
}

func (b *Build) Run(state *State) error {
	return b.RunNode(state, 0)
}

func (b *Build) RunNode(state *State, flags parser.NodeType) error {
	b.flags = flags
	return b.walk(b.tree.Root, state)
}

func (b *Build) walk(node parser.Node, state *State) (err error) {

	switch node := node.(type) {
	case *parser.ListNode:
		for _, node := range node.Nodes {
			err = b.walk(node, state)
			if err != nil {
				break
			}
		}

	case *parser.FilterNode:
		if isMatch(node, state) {
			b.walk(node.Node, state)
		}

	case *parser.DockerNode:
		if shouldSkip(b.flags, node.NodeType) {
			break
		}
		if len(node.Image) == 0 {
			break
		}
		// auth for accessing private docker registries
		var auth *dockerclient.AuthConfig
		// auth to nil if password or token not set
		if len(node.AuthConfig.Password) != 0 || len(node.AuthConfig.RegistryToken) != 0 {
			auth = &dockerclient.AuthConfig{
				Username:      node.AuthConfig.Username,
				Password:      node.AuthConfig.Password,
				Email:         node.AuthConfig.Email,
				RegistryToken: node.AuthConfig.RegistryToken,
			}
		}
		switch node.Type() {

		case parser.NodeBuild:
			// TODO(bradrydzewski) this should be handled by the when block
			// by defaulting the build steps to run when not failure. This is
			// required now that we support multi-build steps.
			if state.Failed() {
				return
			}

			conf := toContainerConfig(node)
			conf.Env = append(conf.Env, toEnv(state)...)
			conf.WorkingDir = state.Workspace.Path
			if state.Repo.IsPrivate {
				script.Encode(state.Workspace, conf, node)
			} else {
				script.Encode(nil, conf, node)
			}

			info, err := docker.Run(state.Client, conf, auth, node.Pull, state.Stdout, state.Stderr)
			if err != nil {
				state.Exit(255)
			} else if info.State.ExitCode != 0 {
				state.Exit(info.State.ExitCode)
			}

		case parser.NodeCompose:
			conf := toContainerConfig(node)
			_, err := docker.Start(state.Client, conf, auth, node.Pull)
			if err != nil {
				state.Exit(255)
			}

		default:
			conf := toContainerConfig(node)
			conf.Cmd = toCommand(state, node)
			info, err := docker.Run(state.Client, conf, auth, node.Pull, state.Stdout, state.Stderr)
			if err != nil {
				state.Exit(255)
			} else if info.State.ExitCode != 0 {
				state.Exit(info.State.ExitCode)
			}
		}
	}

	return nil
}

func expectMatch() {

}

func maybeResolveImage() {}

func maybeEscalate(conf dockerclient.ContainerConfig, node *parser.DockerNode) {
	if node.Image == "plugins/drone-docker" || node.Image == "plugins/drone-gcr" {
		return
	}
	conf.Volumes = nil
	conf.HostConfig.NetworkMode = ""
	conf.HostConfig.Privileged = true
	conf.Entrypoint = []string{}
	conf.Cmd = []string{}
}

// shouldSkip is a helper function that returns true if
// node execution should be skipped. This happens when
// the build is executed for a subset of build steps.
func shouldSkip(flags parser.NodeType, nodeType parser.NodeType) bool {
	return flags != 0 && flags&nodeType == 0
}

// shouldEscalate is a helper function that returns true
// if the plugin should be escalated to start the container
// in privileged mode.
func shouldEscalate(node *parser.DockerNode) bool {
	return node.Image == "plugins/drone-docker" ||
		node.Image == "plugins/drone-gcr"
}
