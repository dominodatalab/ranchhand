package rancher

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

var httpClient = http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

type LoginCredentials struct {
	Username string
	Password string
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type loginResponse struct {
	Token string
}

func Ping(host string) error {
	pingURL, err := url.Parse(fmt.Sprintf("https://%s/ping", host))
	if err != nil {
		return err
	}

	resp, err := httpClient.Get(pingURL.String())
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("rancher ping failed with status (%d)", resp.StatusCode)
	}
	return nil
}

func Login(host string, creds *LoginCredentials) (token string, err error) {
	loginURL, err := url.Parse(fmt.Sprintf("https://%s/v3-public/localProviders/local?action=login", host))
	if err != nil {
		return
	}
	body, err := json.Marshal(creds)
	if err != nil {
		return
	}
	resp, err := httpClient.Post(loginURL.String(), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		body, _ := ioutil.ReadAll(resp.Body)
		err = &authError{string(body)}
	case resp.StatusCode != http.StatusCreated:
		err = errors.Errorf("rancher login failed with status (%d)", resp.StatusCode)
	}
	if err != nil {
		return
	}

	loginData := new(loginResponse)
	if err = json.NewDecoder(resp.Body).Decode(loginData); err != nil {
		return "", errors.Wrap(err, "unexpected rancher response")
	}
	return loginData.Token, nil
}

func ChangePassword(host, token string, input *ChangePasswordInput) error {
	cpURL, err := url.Parse(fmt.Sprintf("https://%s/v3/users?action=changepassword", host))
	if err != nil {
		return err
	}
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", cpURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	bearer := "Bearer " + token
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "rancher password change request failed")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.Errorf("rancher password change returned unexpected status (%d): %v", resp.StatusCode, string(body))
	}
	return nil
}
