package exec

import (
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/drone/drone-exec/docker"
	"github.com/drone/drone-exec/parser"
	"github.com/drone/drone-exec/runner"
	"github.com/drone/drone-exec/yaml"
	"github.com/drone/drone-exec/yaml/inject"
	"github.com/drone/drone-exec/yaml/path"
	"github.com/drone/drone-exec/yaml/secure"
	"github.com/drone/drone-exec/yaml/shasum"
	"github.com/drone/drone-plugin-go/plugin"
	"github.com/samalba/dockerclient"

	log "github.com/Sirupsen/logrus"
)

// Payload defines the raw plugin payload that
// stores the build metadata and configuration.
type Payload struct {
	Yaml      string               `json:"config"`
	YamlEnc   string               `json:"secret"`
	Repo      *plugin.Repo         `json:"repo"`
	Build     *plugin.Build        `json:"build"`
	BuildLast *plugin.Build        `json:"build_last"`
	Job       *plugin.Job          `json:"job"`
	Netrc     []*plugin.NetrcEntry `json:"netrc"`
	Keys      *plugin.Keypair      `json:"keys"`
	System    *plugin.System       `json:"system"`
	Workspace *plugin.Workspace    `json:"workspace"`
}

// Options defines execution options.
type Options struct {
	Cache  bool   // execute cache steps
	Clone  bool   // execute clone steps
	Build  bool   // execute build steps
	Deploy bool   // execute deploy steps
	Notify bool   // execute notify steps
	Debug  bool   // execute in debug mode
	Force  bool   // force pull plugin images
	Mount  string // mounts the volume on the host machine
}

// Error reports an error during execution of a build.
type Error struct {
	ExitCode int // exit code
}

func (e *Error) Error() string { return fmt.Sprintf("build failed (exit code %d)", e.ExitCode) }

// Exec executes a build with the given payload and options. If the
// build fails, an *Error is returned.
func Exec(payload Payload, opt Options, outw, errw io.Writer) error {
	var sec *secure.Secure
	if payload.Keys != nil && len(payload.YamlEnc) != 0 {
		var err error
		sec, err = secure.Parse(payload.YamlEnc, payload.Keys.Private)
		if err != nil {
			return fmt.Errorf("decrypting encrypted secrets: %s", err)
		}
		log.Debugln("Successfully decrypted secrets")
	}

	// TODO This block of code (and the above block) need to be cleaned
	//      up and written in a manner that facilitates better unit testing.
	if sec != nil {
		verified := shasum.Check(payload.Yaml, sec.Checksum)

		// the checksum should be invalidated if the repository is
		// public, and the build is a pull request, and the checksum
		// value was not provided.
		if payload.Build.Event == plugin.EventPull && !payload.Repo.IsPrivate && len(sec.Checksum) == 0 {
			verified = false
		}

		switch {
		case verified && payload.Build.Event == plugin.EventPull:
			log.Debugln("Injected secrets into Yaml safely")
			var err error
			payload.Yaml, err = inject.InjectSafe(payload.Yaml, sec.Environment.Map())
			if err != nil {
				return fmt.Errorf("injecting yaml secrets: %s", err)
			}
		case verified:
			log.Debugln("Injected secrets into Yaml")
			payload.Yaml = inject.Inject(payload.Yaml, sec.Environment.Map())
		case !verified:
			// if we can't validate the Yaml file we don't inject
			// secrets, and therefore shouldn't bother running the
			// deploy and notify tests.
			opt.Deploy = false
			opt.Notify = false
			log.Debugln("Unable to validate Yaml checksum.", sec.Checksum)
		}
	}

	// injects the matrix configuration parameters
	// into the yaml prior to parsing.
	injectParams := map[string]string{
		"COMMIT_SHORT": payload.Build.Commit, // DEPRECATED
		"COMMIT":       payload.Build.Commit,
		"BRANCH":       payload.Build.Branch,
		"BUILD_NUMBER": strconv.Itoa(payload.Build.Number),
	}
	if payload.Build.Event == plugin.EventTag {
		injectParams["TAG"] = strings.TrimPrefix(payload.Build.Ref, "refs/tags/")
	}
	payload.Yaml = inject.Inject(payload.Yaml, payload.Job.Environment)
	payload.Yaml = inject.Inject(payload.Yaml, injectParams)

	// safely inject global variables
	var globals = map[string]string{}
	for _, s := range payload.System.Globals {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) != 2 {
			continue
		}
		globals[parts[0]] = parts[1]
	}
	if payload.Repo.IsPrivate {
		payload.Yaml = inject.Inject(payload.Yaml, globals)
	} else {
		payload.Yaml, _ = inject.InjectSafe(payload.Yaml, globals)
	}

	// extracts the clone path from the yaml. If
	// the clone path doesn't exist it uses a path
	// derrived from the repository uri.
	payload.Workspace = &plugin.Workspace{Keys: payload.Keys, Netrc: payload.Netrc}
	payload.Workspace.Path = path.Parse(payload.Yaml, payload.Repo.Link)
	payload.Workspace.Root = "/drone/src"
	log.Debugf("Using workspace %s", payload.Workspace.Path)

	rules := []parser.RuleFunc{
		parser.ImageName,
		parser.ImageMatchFunc(payload.System.Plugins),
		parser.ImagePullFunc(opt.Force),
		parser.SanitizeFunc(payload.Repo.IsTrusted), //&& !plugin.PullRequest(payload.Build)
		parser.CacheFunc(payload.Repo.FullName),
		parser.DebugFunc(yaml.ParseDebugString(payload.Yaml)),
		parser.Escalate,
		parser.HttpProxy,
		parser.DefaultNotifyFilter,
	}
	if len(opt.Mount) != 0 {
		log.Debugf("Mounting %s as workspace %s",
			opt.Mount,
			payload.Workspace.Path,
		)
		rules = append(rules, parser.MountFunc(
			opt.Mount,
			payload.Workspace.Path,
		))
	}
	tree, err := parser.Parse(payload.Yaml, rules)
	if err != nil {
		// TODO(sqs): There was a comment here saying "print error
		// messages in debug mode only". Is this because of security
		// (e.g., the decrypted YAML secrets could leak in the error
		// message)? If so, don't return the err here; instead, return
		// a simple error message such as "error parsing yaml".
		return err
	}
	r := runner.Load(tree)

	// TODO(sqs!native-ci): copied temporarily from https://github.com/drone/drone-exec/pull/13, godep update when that is merged into drone-exec
	daemonURL := os.Getenv("DOCKER_HOST")
	if daemonURL == "" {
		daemonURL = "unix:///var/run/docker.sock"
	}
	var tlsConfig *tls.Config
	if path := os.Getenv("DOCKER_CERT_PATH"); os.Getenv("DOCKER_TLS_VERIFY") != "" && path != "" {
		tlsConfig, err = TLSConfigFromCertPath(path)
		if err != nil {
			return err
		}
	}
	client, err := dockerclient.NewDockerClient(daemonURL, tlsConfig)
	if err != nil {
		return err
	}

	// // creates a wrapper Docker client that uses an ambassador
	// // container to create a pod-like environment.
	controller, err := docker.NewClient(client)
	if err != nil {
		return fmt.Errorf("creating docker ambassador container: %s", err)
	}
	defer controller.Destroy()

	// watch for sigkill (timeout or cancel build)
	killc := make(chan os.Signal, 1)
	signal.Notify(killc, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-killc
		log.Println("Cancel request received, killing process")
		controller.Destroy() // possibe race here. implement lock on the other end
		os.Exit(130)         // cancel is treated like ctrl+c
	}()

	go func() {
		var timeout = payload.Repo.Timeout
		if timeout == 0 {
			timeout = 60
		}
		<-time.After(time.Duration(timeout) * time.Minute)
		log.Println("Timeout request received, killing process")
		controller.Destroy() // possibe race here. implement lock on the other end
		os.Exit(128)         // cancel is treated like ctrl+c
	}()

	state := &runner.State{
		Client:    controller,
		Stdout:    outw,
		Stderr:    errw,
		Repo:      payload.Repo,
		Build:     payload.Build,
		BuildLast: payload.BuildLast,
		Job:       payload.Job,
		System:    payload.System,
		Workspace: payload.Workspace,
	}
	if opt.Cache {
		log.Debugln("Running Cache step")
		err = r.RunNode(state, parser.NodeCache)
		if err != nil {
			log.Debugln(err)
		}
	}
	if opt.Clone {
		log.Debugln("Running Clone step")
		err = r.RunNode(state, parser.NodeClone)
		if err != nil {
			log.Debugln(err)
		}
	}
	if opt.Build && !state.Failed() {
		log.Debugln("Running Build and Compose steps")
		if err := r.RunNode(state, parser.NodeCompose); err != nil {
			log.Debugln(err)
		}
		if err := r.RunNode(state, parser.NodeBuild); err != nil {
			log.Debugln(err)
		}
	}
	if opt.Deploy && !state.Failed() {
		log.Debugln("Running Publish and Deploy steps")
		if err := r.RunNode(state, parser.NodePublish); err != nil {
			log.Debugln(err)
		}
		if err := r.RunNode(state, parser.NodeDeploy); err != nil {
			log.Debugln(err)
		}
	}

	// if the build is not failed, at this point
	// we can mark as successful
	if !state.Failed() {
		state.Job.Status = plugin.StateSuccess
		state.Build.Status = plugin.StateSuccess
	}

	if opt.Cache {
		log.Debugln("Running post-Build Cache steps")
		err = r.RunNode(state, parser.NodeCache)
		if err != nil {
			log.Debugln(err)
		}
	}
	if opt.Notify {
		log.Debugln("Running Notify steps")
		err = r.RunNode(state, parser.NodeNotify)
		if err != nil {
			log.Debugln(err)
		}
	}

	if state.Failed() {
		controller.Destroy()
		return &Error{ExitCode: state.ExitCode()}
	}

	return nil
}
