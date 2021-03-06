package web

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/G-Node/gin-cli/util"
	gogs "github.com/gogits/go-gogs-client"
)

// UserToken struct for username and token
type UserToken struct {
	Username string
	Token    string
}

// Client struct for making requests
type Client struct {
	Host string
	UserToken
	web *http.Client
}

func urlJoin(parts ...string) string {
	// First part must be a valid URL
	u, err := url.Parse(parts[0])
	util.CheckErrorMsg(err, "Bad URL in urlJoin")

	for _, part := range parts[1:] {
		u.Path = path.Join(u.Path, part)
	}
	return u.String()
}

// Get sends a GET request to address.
// The address is appended to the client host, so it should be specified without a host prefix.
func (cl *Client) Get(address string) (*http.Response, error) {
	requrl := urlJoin(cl.Host, address)
	req, err := http.NewRequest("GET", requrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/jsonAuthorization")
	util.LogWrite("Performing GET with token: %s", cl.Token)
	if cl.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", cl.Token))
		util.LogWrite("Added token to GET")
	}
	util.LogWrite("Performing GET: %s", req.URL)
	return cl.web.Do(req)
}

// Post sends a POST request to address with the provided data.
// The address is appended to the client host, so it should be specified without a host prefix.
func (cl *Client) Post(address string, data interface{}) (*http.Response, error) {
	datajson, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	requrl := urlJoin(cl.Host, address)
	req, err := http.NewRequest("POST", requrl, bytes.NewReader(datajson))
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/jsonAuthorization")
	if cl.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", cl.Token))
		util.LogWrite("Added token to POST")
	}
	util.LogWrite("Performing POST: %s", req.URL)
	return cl.web.Do(req)
}

// PostBasicAuth sends a POST request to address with the provided data.
// The username and password are used to perform Basic authentication.
func (cl *Client) PostBasicAuth(address, username, password string, data interface{}) (*http.Response, error) {
	datajson, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	requrl := urlJoin(cl.Host, address)
	req, err := http.NewRequest("POST", requrl, bytes.NewReader(datajson))
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", gogs.BasicAuthEncode(username, password)))
	util.LogWrite("Performing POST: %s", req.URL)
	return cl.web.Do(req)
}

// Delete sends a DELETE request to address.
func (cl *Client) Delete(address string) (*http.Response, error) {
	requrl := urlJoin(cl.Host, address)
	req, err := http.NewRequest("DELETE", requrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/jsonAuthorization")
	if cl.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", cl.Token))
		util.LogWrite("Added token to DELETE")
	}
	util.LogWrite("Performing DELETE: %s", req.URL)
	return cl.web.Do(req)
}

// NewClient creates a new client for a given host.
func NewClient(host string) *Client {
	return &Client{Host: host, web: &http.Client{}}
}

// LoadToken reads the username and auth token from the token file and sets the
// values in the struct.
func (ut *UserToken) LoadToken() error {
	// TODO: Don't reload if already set
	util.LogWrite("Loading token")
	path, err := util.ConfigPath(false)
	if err != nil {
		return fmt.Errorf("Could not read token: Error accessing config directory.")
	}
	filepath := filepath.Join(path, "token")
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("Error loading user token")
	}
	defer closeFile(file)

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(ut)
	if err != nil {
		return err
	}
	util.LogWrite("Token read OK")
	return nil
}

// StoreToken saves the username and auth token to the token file.
func (ut *UserToken) StoreToken() error {
	util.LogWrite("Saving token. ")
	path, err := util.ConfigPath(true)
	if err != nil {
		return fmt.Errorf("Could not save token: Error creating or accessing config directory.")
	}
	filepath := filepath.Join(path, "token")
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("Error saving user token.")
	}
	defer closeFile(file)

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(ut)
	if err != nil {
		return err
	}
	util.LogWrite("Saved")
	return nil
}

// DeleteToken deletes the token file if it exists. It essentially logs out the user.
func DeleteToken() error {
	path, err := util.ConfigPath(false)
	if err != nil {
		return fmt.Errorf("Could not delete token: Error accessing config directory.")
	}
	filepath := filepath.Join(path, "token")
	err = os.Remove(filepath)
	if err != nil {
		return err
	}
	util.LogWrite("Token deleted")
	return nil
}

// CloseRes closes a given result buffer (for use with defer).
func CloseRes(b io.ReadCloser) {
	err := b.Close()
	util.CheckErrorMsg(err, "Error during cleanup.")
}

func closeFile(f *os.File) {
	_ = f.Close()
}
