package ranchhand

import (
	"io"
	"net/http"
	"os"

	"github.com/Masterminds/semver"
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
