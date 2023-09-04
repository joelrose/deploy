package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

func TestIntegration(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	err = pool.Client.Ping()
	require.NoError(t, err)

	resource, err := pool.BuildAndRunWithOptions(
		"./testdata/Dockerfile.dind_ssh",
		&dockertest.RunOptions{
			Name:         "dind-ssh-image",
			Privileged:   true,
			ExposedPorts: []string{"22/tcp", "80/tcp"},
		},
	)
	require.NoError(t, err)

	if err = pool.Retry(func() error {
		conn, err := net.DialTimeout("tcp", "localhost:"+resource.GetPort("22/tcp"), 2*time.Second)
		if err != nil {
			return err //nolint:wrapcheck
		}
		conn.Close()

		return nil
	}); err != nil {
		log.Fatalf("Failed to connect to container: %v", err)
	}

	privateKey, err := os.ReadFile("./testdata/devops_key.pem")
	require.NoError(t, err)

	t.Setenv("SSH_PRIVATE_KEY", string(privateKey))

	os.Args = []string{
		"deploy",
		"-environment=test",
		"-sshUsername=devops",
		"-registryUsername=joelrose",
		"-sshPort=" + resource.GetPort("22/tcp"),
		"-path=testdata",
	}

	time.Sleep(1 * time.Second)

	main()

	time.Sleep(1 * time.Second)

	httpPort := resource.GetPort("80/tcp")
	res, err := http.Get("http://localhost:" + httpPort) //nolint:noctx
	require.NoError(t, err)

	require.NoError(t, res.Body.Close())

	require.Equal(t, http.StatusOK, res.StatusCode)

	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}
