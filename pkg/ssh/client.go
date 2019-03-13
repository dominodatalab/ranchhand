package ssh

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// todo: should we add retry logic here?
const Timeout = 5 * time.Second

type Client struct {
	addr  string
	inner *ssh.Client
}

func Connect(host string, port uint, user, sshKeyPath string) (*Client, error) {
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

	sockAddr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", sockAddr, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial host")
	}

	return &Client{addr: host, inner: client}, nil
}

func (c *Client) RemoteAddr() string {
	return c.addr
}

func (c *Client) ExecuteCmd(cmd string) (string, error) {
	session, err := c.inner.NewSession()
	if err != nil {
		return "", errors.Wrap(err, "failed to create new session")
	}
	defer session.Close()

	buffer, err := session.CombinedOutput(cmd)
	output := string(buffer)
	if err != nil {
		return "", errors.Wrapf(err, "failed to run remote command: %s", output)
	}

	return strings.TrimSpace(output), nil
}
