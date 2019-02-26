package ssh

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

const Timeout = 5 * time.Second

type Client struct {
	inner *ssh.Client
}

func Connect(host, user, sshKeyPath string) (*Client, error) {
	buffer, err := ioutil.ReadFile(sshKeyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, err
	}
	certAuth := ssh.PublicKeys(signer)

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{certAuth},
		Timeout:         Timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial host")
	}

	return &Client{inner: client}, nil
}

func (c *Client) ExecuteCmd(cmd string) (string, error) {
	session, err := c.inner.NewSession()
	if err != nil {
		return "", errors.Wrap(err, "failed to create new session")
	}
	defer session.Close()

	buffer, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", errors.Wrap(err, "failed to run remote command")
	}

	return strings.TrimSpace(string(buffer)), nil
}
