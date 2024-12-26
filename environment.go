package flargs

import (
	"bytes"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// Enviroment is an execution environment for a Command.
// In the context of a CLI, these would be [os.StdIn], [os.StdOut], etc.
// In the context of a test-suite, you can use [bytes.Buffer] and [fstest.MapFS].
// For benchmarking, you can use a [NullDevice].
type Environment struct {
	InputStream  io.ReadWriter
	OutputStream io.ReadWriter
	ErrorStream  io.ReadWriter
	Randomness   rand.Source
	Filesystem   afero.Fs
	Variables    map[string]string
	Arguments    []string
}

func (e Environment) GetOutput() []byte {
	buf, _ := io.ReadAll(e.OutputStream)
	return buf
}

func (e Environment) GetError() []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(e.ErrorStream)
	return buf.Bytes()
}

func (e Environment) GetInput() []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(e.InputStream)
	return buf.Bytes()
}

// NewCLIEnvironment produces an Environment suitable for a CLI.
// It's a helper function with sane defaults.
func NewCLIEnvironment(baseDir string) *Environment {

	envAsMap := func(envs []string) map[string]string {
		m := make(map[string]string)
		i := 0
		for _, s := range envs {
			i = strings.IndexByte(s, '=')
			m[s[0:i]] = s[i+1:]
		}
		return m
	}
	//	import parent env vars
	vars := envAsMap(os.Environ())
	vars["FLARGS_EXE_ENVIRONMENT"] = "cli"

	realFs := afero.NewOsFs()

	env := Environment{
		InputStream:  os.Stdin,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stderr,
		Randomness:   rand.NewSource(time.Now().UnixNano()),
		Filesystem:   realFs,
		Variables:    vars,
		Arguments:    os.Args,
	}
	return &env
}

// NewTestingEnvironment produces an [Environment] suitable for testing.
// Pass in a "randSource" that offers a level of determinism that works for you.
// For good ole fashioned regular randomness, pass in [rand.Reader]
// If your program doesn't use randomness, just pass in nil.
func NewTestingEnvironment(randSource rand.Source) *Environment {
	env := Environment{
		InputStream:  new(bytes.Buffer),
		OutputStream: new(bytes.Buffer),
		ErrorStream:  new(bytes.Buffer),
		Randomness:   randSource,
		Filesystem:   afero.NewMemMapFs(),
		Variables: map[string]string{
			"FLARGS_EXE_ENVIRONMENT": "testing",
		},
		Arguments: []string{},
	}
	return &env
}

// a NullDevice satisfies necesary interfaces but drops all information on the floor
type NullDevice struct {
	io.Writer
	afero.Fs
}

func (b NullDevice) Read(_ []byte) (int, error) {
	return 0, nil
}

func NewNullEnvironment() *Environment {
	e := Environment{
		InputStream:  NullDevice{io.Discard, afero.NewMemMapFs()},
		OutputStream: NullDevice{io.Discard, afero.NewMemMapFs()},
		ErrorStream:  NullDevice{io.Discard, afero.NewMemMapFs()},
		Filesystem:   afero.NewMemMapFs(),
		Variables:    map[string]string{},
		Arguments:    []string{},
	}
	return &e
}
