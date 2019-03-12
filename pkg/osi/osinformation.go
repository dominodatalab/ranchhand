package osi

import "strings"

const (
	UbuntuOS = "ubuntu"
	CentOS   = "centos"
	RHELOS   = "rhel"
)

type Info struct {
	ID         string
	VersionID  string
	PrettyName string
}

func Parse(s string) *Info {
	kvPairs := strings.Split(s, "\n")
	info := make(map[string]string)
	for _, pair := range kvPairs {
		if len(pair) > 0 {
			z := strings.Split(pair, "=")
			info[z[0]] = strings.Trim(z[1], "\"")
		}
	}

	return &Info{
		ID:         info["ID"],
		VersionID:  info["VERSION_ID"],
		PrettyName: info["PRETTY_NAME"],
	}
}

func (i *Info) IsUbuntu() bool {
	return i.ID == UbuntuOS
}

func (i *Info) IsCentOS() bool {
	return i.ID == CentOS
}

func (i *Info) IsRHEL() bool {
	return i.ID == RHELOS
}
