package ssh

import (
        "fmt"
	"time"

        "github.com/appleboy/easyssh-proxy"
)

const Timeout = 5 * time.Second

func Connect(hostname, sshUser, proxyHost, proxyUser, sshKeyPath string, sshPort uint) (*easyssh.MakeConfig, error) {
    port := fmt.Sprint(sshPort)
    ssh := &easyssh.MakeConfig{
        User: sshUser,
        Server: hostname,
        Port: port,
        KeyPath: sshKeyPath,
        Timeout: Timeout,
    }

    if proxyHost != "" {
        ssh.Proxy = easyssh.DefaultConfig{
            User: proxyUser,
            Server: proxyHost,
            Port: port,
            KeyPath: sshKeyPath,
        }
    }

    return ssh, nil
}
