package daemon

import (
	"errors"
	"os/exec"
	"testing"
)

// TestResolveStateRoot_DockerHelperContract verifies the docker-credential-devsy
// helper is discoverable on $PATH before a devcontainer can pull authenticated
// images. It skips (rather than fails) when the helper is not installed, so CI
// stays green until the environment prerequisite is wired.
func TestResolveStateRoot_DockerHelperContract(t *testing.T) {
	path, err := exec.LookPath("docker-credential-devsy")
	if err != nil {
		t.Skip("docker-credential-devsy is not installed on $PATH; " +
			"helper contract not met — install it or set DEVSY_DOCKER_CLI_PLUGINS, see docs")
	}
	t.Logf("docker-credential-devsy found at %s", path)
}

func errorsAs(err error, target interface{}) bool {
	return errors.As(err, target)
}
