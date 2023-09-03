package main

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

type SSH struct {
	config *ssh.ClientConfig
}

func NewSSH(username, privateKey string) (*SSH, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %v", err)
	}

	sshConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO(joelrose): this is insecure
	}

	return &SSH{sshConfig}, nil
}

func (s *SSH) RunCommand(addr string, command string) (string, error) {
	client, err := ssh.Dial("tcp", addr, s.config)
	if err != nil {
		return "", fmt.Errorf("dialing host: %v, with error: %v", addr, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("creating session: %v", err)
	}
	defer session.Close()

	out, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("running command: %v, with error: %v, out: %v", command, err, out)
	}

	return string(out), nil
}
