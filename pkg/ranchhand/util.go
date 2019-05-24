package ranchhand

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/dominodatalab/ranchhand/pkg/ssh"
	"github.com/pkg/errors"
)

func ensureDirectory(dir string) error {
	if _, serr := os.Stat(dir); os.IsNotExist(serr) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func ensureRemoteDirectory(client *ssh.Client, name string) error {
	_, err := client.ExecuteCmd(fmt.Sprintf("sudo mkdir -p %s", name))
	return err
}

func downloadFile(filepath, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, resp.Body)
	return err
}

func constrainVersion(constraint, version string) error {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return err
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	if ok, errs := c.Validate(v); !ok {
		return errors.Errorf("version validation failed: %v", errs)
	}

	return nil
}

func BuildNodes(nodeIPs []string) (nodes []Node) {
	for _, compIP := range nodeIPs {
		bits := strings.Split(compIP, ":")

		node := Node{PublicIP: bits[0]}
		if len(bits) == 2 {
			node.PrivateIP = bits[1]
		}
		nodes = append(nodes, node)
	}

	return nodes
}
