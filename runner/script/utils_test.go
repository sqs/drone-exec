package script

import (
	"testing"

	"github.com/franela/goblin"
)

func Test_Utils(t *testing.T) {

	g := goblin.Goblin(t)
	g.Describe("Script writer", func() {

		g.It("Should encode scripts", func() {
			want := "echo ZWNobyBmb28= | base64 -d | /bin/sh"
			got := encode([]byte("echo foo"))
			g.Assert(want).Equal(got)
		})

		g.It("Should trace one command", func() {
			got := trace("echo hello")
			g.Assert(traced).Equal(got)
		})

		g.It("Should trace many commands", func() {
			cmds := []string{"echo hello", "echo goodbye"}
			got := writeCmds(cmds)
			g.Assert(got).Equal(script)
		})

		g.It("Should generate netrc script", func() {
			got := writeNetrc("foo", "bar", "baz")
			g.Assert(netrc).Equal(got)
		})

		g.It("Should generate empty netrc script", func() {
			got := writeNetrc("", "bar", "baz")
			g.Assert(got).Equal("")
		})

		g.It("Should generate ssh key script", func() {
			got := writeKey("-----BEGIN RSA PRIVATE KEY----- MIIEpQIBAAKCAQEA3Tz2...")
			g.Assert(keys).Equal(got)
		})

		g.It("Should generate empty ssh key script", func() {
			got := writeKey("")
			g.Assert(got).Equal("")
		})
	})
}

var netrc = `
cat <<EOF > $HOME/.netrc
machine foo
login bar
password baz
EOF
chmod 0600 $HOME/.netrc
`

var keys = `
mkdir -p -m 0700 $HOME/.ssh
cat <<EOF > $HOME/.ssh/id_rsa
-----BEGIN RSA PRIVATE KEY----- MIIEpQIBAAKCAQEA3Tz2...
EOF
chmod 0600 $HOME/.ssh/id_rsa

cat <<EOF > $HOME/.ssh/config
StrictHostKeyChecking no
EOF
`

var script = `
echo JCBlY2hvIGhlbGxvCg== | base64 -d
echo hello

echo JCBlY2hvIGdvb2RieWUK | base64 -d
echo goodbye
`

var traced = `
echo JCBlY2hvIGhlbGxvCg== | base64 -d
echo hello
`
