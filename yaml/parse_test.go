package yaml

import (
	"testing"

	"github.com/franela/goblin"
)

func TestParse(t *testing.T) {

	conf, err := ParseString(sample)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	g := goblin.Goblin(t)
	g.Describe("Parse Yaml", func() {

		g.It("Should parse images", func() {
			g.Assert(conf.Clone.Image).Equal("git")
		})

		g.It("Should parse image force-pull", func() {
			g.Assert(conf.Clone.Pull).Equal(true)
		})

		g.It("Should parse variable arguments", func() {
			g.Assert(conf.Clone.Vargs["path"]).Equal("github.com/octocat/hello-world")
		})

		g.It("Should parse build image", func() {
			g.Assert(conf.Build[0].Image).Equal("golang")
		})

		g.It("Should parse build commands", func() {
			g.Assert(conf.Build[0].Commands).Equal([]string{"go build", "go test"})
		})

		g.It("Should parse volume configuration", func() {
			g.Assert(conf.Build[0].Volumes).Equal([]string{"/tmp/volumes"})
		})

		g.It("Should parse network configuration", func() {
			g.Assert(conf.Build[0].Net).Equal("bridge")
		})

		g.It("Should parse environment variable map", func() {
			g.Assert(conf.Clone.Environment).Equal(
				MapEqualSlice{"GIT_DIR=.git"},
			)
		})

		g.It("Should parse environment variable slice", func() {
			g.Assert(conf.Build[0].Environment).Equal(
				MapEqualSlice{"GO15VENDOREXPERIMENT=1"},
			)
		})

		g.It("Should parse docker command slice", func() {
			g.Assert(conf.Compose[0].Command).Equal(
				Command{
					"redis-server",
					"/usr/local/etc/redis/redis.conf",
					"--appendonly",
					"yes",
				},
			)
		})

		g.It("Should parse docker command string", func() {
			g.Assert(conf.Compose[1].Command).Equal(
				Command{
					"--storageEngine",
					"wiredTiger",
				},
			)
		})

		g.It("Should allow multiple plugins of same type", func() {
			s := conf.Deploy
			g.Assert(s[0].Image).Equal("heroku")
			g.Assert(s[1].Image).Equal("heroku")
		})

		g.It("Should maintain plugin ordering", func() {
			s := conf.Deploy
			g.Assert(s[0].Vargs["app"]).Equal("foo.com")
			g.Assert(s[1].Vargs["app"]).Equal("dev.foo.com")
		})

		g.It("Should parse plugin filters", func() {
			s := conf.Deploy
			g.Assert(s[0].Filter.Branch).Equal(Stringorslice{"master"})
			g.Assert(s[1].Filter.Repo).Equal("octocat/helloworld")
			g.Assert(s[1].Filter.Matrix).Equal(map[string]string{"go_version": "1.5"})
		})

		g.It("Should error when Yaml is malformed", func() {
			_, err := ParseString(malformed)
			g.Assert(err.Error()).Equal("yaml: found unexpected ':'")
		})

		g.It("Should parse a Yaml with variables", func() {
			conf, err := ParseString(variables)
			g.Assert(err).Equal(nil)
			g.Assert(conf.Build[0].Image).Equal("golang")
			g.Assert(conf.Build[0].Environment).Equal(
				MapEqualSlice{"GO15VENDOREXPERIMENT=1"},
			)
			g.Assert(conf.Build[0].Commands).Equal([]string{"go build", "go test"})
			g.Assert(conf.Build[0].Volumes).Equal([]string{"/tmp/volumes"})
			g.Assert(conf.Build[0].Net).Equal("bridge")
			g.Assert(conf.Build[0].Privileged).Equal(true)
		})

		g.It("Should parse a Yaml with multiple build steps", func() {
			conf, err := ParseString(multiBuild)
			g.Assert(err).Equal(nil)
			g.Assert(conf.Build[0].Image).Equal("golang")
			g.Assert(conf.Build[0].Environment).Equal(
				MapEqualSlice{"GO15VENDOREXPERIMENT=1"},
			)
			g.Assert(conf.Build[0].Commands).Equal([]string{"go build", "go test"})

			g.Assert(conf.Build[1].Image).Equal("node")
			g.Assert(conf.Build[1].Commands).Equal([]string{"npm install", "npm test"})
		})
	})
}

var sample = `
clone:
  image: git
  pull: true
  path: github.com/octocat/hello-world
  environment:
    GIT_DIR: .git

build:
  image: golang
  environment:
    - GO15VENDOREXPERIMENT=1
  commands:
    - go build
    - go test
  volumes:
    - /tmp/volumes
  net: bridge
  privileged: true
  auth_config:
    password: test
    username: test
    email: test@example.com

compose:
  redis:
    image: library/redis
    command: redis-server /usr/local/etc/redis/redis.conf --appendonly yes

  mongo:
    image: library/mongo
    command:
      - --storageEngine
      - wiredTiger

deploy:
  heroku:
    app: foo.com
    when:
      branch: master
  heroku:
    app: dev.foo.com
    when:
      repo: octocat/helloworld
      branch: somebranch
      matrix:
        go_version: 1.5
`

var multiBuild = `
build:
  backend:
    image: golang
    environment:
      - GO15VENDOREXPERIMENT=1
    commands:
      - go build
      - go test
  frontent:
    image: node
    commands:
      - npm install
      - npm test
`

var malformed = `build: { image: golang:1.4.2, commands: [ go build, go test ] }`

var variables = `
build_values: &BUILD_VALUES
  image: golang
  environment:
    - GO15VENDOREXPERIMENT=1
  commands:
    - go build
    - go test
  volumes:
    - /tmp/volumes
  net: bridge

build:
  <<: *BUILD_VALUES
  privileged: true
`
