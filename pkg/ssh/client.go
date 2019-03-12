package ssh

import (
	"fmt"
	"time"

	"github.com/appleboy/easyssh-proxy"
)

const Timeout = 5 * time.Second

type Client struct {
	inner *easyssh.MakeConfig
}

func Connect(hostname, sshUser, sshKeyPath string, sshPort uint) *Client {
	port := fmt.Sprint(sshPort)
	ssh := &easyssh.MakeConfig{
		User:    sshUser,
		Server:  hostname,
		Port:    port,
		KeyPath: sshKeyPath,
		Timeout: Timeout,
	}

	return &Client{inner: ssh}
}

func (c *Client) ExecuteCmd(cmd string) (string, string, bool, error) {
	return (c.inner.Run(cmd, 300*time.Second))
}
