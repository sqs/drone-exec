package script

import (
	"testing"

	"github.com/drone/drone-exec/parser"
	"github.com/drone/drone-plugin-go/plugin"
	"github.com/franela/goblin"
	"github.com/samalba/dockerclient"
)

func Test_Rule(t *testing.T) {

	g := goblin.Goblin(t)
	g.Describe("Script encoder", func() {

		g.It("Should encode the build script", func() {
			c := &dockerclient.ContainerConfig{}
			n := &parser.DockerNode{
				Commands: []string{"go build", "go test"},
			}
			Encode(nil, c, n)
			want := encode(decoded1)
			got := c.Cmd[0]

			g.Assert(c.Entrypoint).Equal(entrypoint)
			g.Assert(want).Equal(got)
		})

		g.It("Should encode the build script with private key and netrc", func() {
			c := &dockerclient.ContainerConfig{}
			n := &parser.DockerNode{
				Commands: []string{"go build", "go test"},
			}
			w := &plugin.Workspace{
				Netrc: []*plugin.NetrcEntry{
					{
						Machine:  "foo",
						Login:    "bar",
						Password: "baz",
					},
					{
						Machine:  "qux",
						Login:    "zip",
						Password: "zap",
					},
				},
				Keys: &plugin.Keypair{
					Private: "-----BEGIN RSA PRIVATE KEY----- MIIEpQIBAAKCAQEA3Tz2...",
				},
			}
			Encode(w, c, n)
			want := encode(decoded2)
			got := c.Cmd[0]

			g.Assert(c.Entrypoint).Equal(entrypoint)
			g.Assert(want).Equal(got)
		})
	})
}

var decoded1 = []byte(`
[ -z "$HOME"  ] && export HOME="/root"
[ -z "$SHELL" ] && export SHELL="/bin/sh"

export GOBIN=/drone/bin
export GOPATH=/drone
export PATH=$PATH:$GOBIN

set -e

echo JCBnbyBidWlsZAo= | base64 -d
go build

echo JCBnbyB0ZXN0Cg== | base64 -d
go test

rm -rf $HOME/.netrc
rm -rf $HOME/.ssh/id_rsa
`)

var decoded2 = []byte(`
[ -z "$HOME"  ] && export HOME="/root"
[ -z "$SHELL" ] && export SHELL="/bin/sh"

export GOBIN=/drone/bin
export GOPATH=/drone
export PATH=$PATH:$GOBIN

set -e

mkdir -p -m 0700 $HOME/.ssh
cat <<EOF > $HOME/.ssh/id_rsa
-----BEGIN RSA PRIVATE KEY----- MIIEpQIBAAKCAQEA3Tz2...
EOF
chmod 0600 $HOME/.ssh/id_rsa

cat <<EOF > $HOME/.ssh/config
StrictHostKeyChecking no
EOF

mkdir -p /etc/apt/apt.conf.d
cat <<EOF > /etc/apt/apt.conf.d/90forceyes
APT::Get::Assume-Yes "true";APT::Get::force-yes "true";
EOF

cat <<EOF >> $HOME/.netrc
machine foo
login bar
password baz

EOF
chmod 0600 $HOME/.netrc

cat <<EOF >> $HOME/.netrc
machine qux
login zip
password zap

EOF
chmod 0600 $HOME/.netrc

echo JCBnbyBidWlsZAo= | base64 -d
go build

echo JCBnbyB0ZXN0Cg== | base64 -d
go test

rm -rf $HOME/.netrc
rm -rf $HOME/.ssh/id_rsa
`)
