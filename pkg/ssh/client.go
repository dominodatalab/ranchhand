package ssh

import (
        "fmt"
	"time"

        "github.com/appleboy/easyssh-proxy"
)

const Timeout = 5 * time.Second

func Connect(hostname, sshUser, sshKeyPath string, sshPort uint) (*easyssh.MakeConfig, error) {
    port := fmt.Sprint(sshPort)
    ssh := &easyssh.MakeConfig{
        User: sshUser,
        Server: hostname,
        Port: port,
        KeyPath: sshKeyPath,
        Timeout: Timeout,
    }

    return ssh, nil
}
