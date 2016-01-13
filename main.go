package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/net/context"

	"github.com/drone/drone-exec/exec"
	"github.com/drone/drone-exec/yaml"
	"github.com/drone/drone-plugin-go/plugin"

	log "github.com/Sirupsen/logrus"
)

func main() {
	var opt exec.Options

	// parses command line flags
	flag.BoolVar(&opt.Cache, "cache", false, "")
	flag.BoolVar(&opt.Clone, "clone", false, "")
	flag.BoolVar(&opt.Build, "build", false, "")
	flag.BoolVar(&opt.Deploy, "deploy", false, "")
	flag.BoolVar(&opt.Notify, "notify", false, "")
	flag.BoolVar(&opt.Debug, "debug", false, "")
	flag.BoolVar(&opt.Force, "pull", false, "")
	flag.StringVar(&opt.Mount, "mount", "", "")
	flag.Parse()

	// unmarshal the json payload via stdin or
	// via the command line args (whichever was used)
	var payload exec.Payload
	if err := plugin.MustUnmarshal(&payload); err != nil {
		log.Fatalln(err)
	}

	// configure the default log format and
	// log levels
	debugFlag := yaml.ParseDebugString(payload.Yaml)
	if debugFlag {
		log.SetLevel(log.DebugLevel)
	}
	log.SetFormatter(new(formatter))

	ctx := context.Background()

	// Watch for sigkill (timeout or cancel build).
	killc := make(chan os.Signal, 1)
	signal.Notify(killc, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-killc
		log.Println("Cancel request received, killing process")
		cancel()
		os.Exit(130)
	}()

	err := exec.Exec(ctx, payload, opt)
	if err != nil {
		log.Println(err)
		switch err := err.(type) {
		case *exec.Error:
			os.Exit(err.ExitCode)
		}
		os.Exit(1)
	}
}

type formatter struct{}

func (f *formatter) Format(entry *log.Entry) ([]byte, error) {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "[%s] %s\n", entry.Level.String(), entry.Message)
	return buf.Bytes(), nil
}
