package cluster

import (
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

var (
	// Binary Parameters
	makefileDir string = "/home/user/fuel-sequencer" // Absolute path to the directory where makefile is located
	wantArch    string = "linux-amd64"               // Arch specified from build binary suffix
	buildPath   string = makefileDir + "/build"      // Path where binary will be built

	// Sequencer Parameters

	// Remote Parameters
	destinations = []destination{
		{host: "localhost", user: "benchmarks", auth: ssh.Password("password"), dir: "/home/benchmarks"},
	}
)

func homeDir(d destination) string {
	return filepath.Join("/home", d.user, ".fuelsequencer")
}
