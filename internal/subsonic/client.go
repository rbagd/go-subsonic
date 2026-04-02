package subsonic

import (
	"context"
	"crypto/md5"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

const (
	ClientName = "go-subsonic"
	APIVersion = "1.16.1"
)

type Client struct {
	BaseURL    string
	Username   string
	Password   string // Stored to generate tokens
	HTTPClient *http.Client
}

var generateSalt = generateRandomString

func NewClient(baseURL, username, password string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Username:   username,
		Password:   password,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Response is a generic wrapper for Subsonic JSON responses
type Response struct {
	SubsonicResponse SubsonicResponse `json:"subsonic-response"`
}

type SubsonicResponse struct {
	Status     string      `json:"status"` // "ok" or "failed"
	Version    string      `json:"version"`
	Error      *APIError   `json:"error,omitempty"`
	Indexes    *Indexes    `json:"indexes,omitempty"`
	Artists    *ArtistsID3 `json:"artists,omitempty"`
	AlbumList2 *AlbumList2 `json:"albumList2,omitempty"`
	Directory  *Directory  `json:"directory,omitempty"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) Ping(ctx context.Context) error {
	req, err := c.newRequest(ctx, "GET", "ping", nil)
	if err != nil {
		return err
	}

	var resp Response
	if err := c.do(req, &resp); err != nil {
		return err
	}

	if resp.SubsonicResponse.Status == "failed" {
		if resp.SubsonicResponse.Error != nil {
			return fmt.Errorf("api error %d: %s", resp.SubsonicResponse.Error.Code, resp.SubsonicResponse.Error.Message)
		}
		return fmt.Errorf("api returned failed status")
	}

	return nil
}

func (c *Client) GetIndexes(ctx context.Context) (*Indexes, error) {
	req, err := c.newRequest(ctx, "GET", "getIndexes", nil)
	if err != nil {
		return nil, err
	}

	var resp Response
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	if err := c.checkError(resp.SubsonicResponse); err != nil {
		return nil, err
	}

	return resp.SubsonicResponse.Indexes, nil
}

func (c *Client) GetArtists(ctx context.Context) (*ArtistsID3, error) {
	req, err := c.newRequest(ctx, "GET", "getArtists", nil)
	if err != nil {
		return nil, err
	}

	var resp Response
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	if err := c.checkError(resp.SubsonicResponse); err != nil {
		return nil, err
	}

	return resp.SubsonicResponse.Artists, nil
}

func (c *Client) GetMusicDirectory(ctx context.Context, id string) (*Directory, error) {
	req, err := c.newRequest(ctx, "GET", "getMusicDirectory", map[string]string{"id": id})
	if err != nil {
		return nil, err
	}

	var resp Response
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	if err := c.checkError(resp.SubsonicResponse); err != nil {
		return nil, err
	}

	return resp.SubsonicResponse.Directory, nil
}

func (c *Client) GetAlbumList2(ctx context.Context, listType string, size, offset int) (*AlbumList2, error) {
	params := map[string]string{"type": listType}
	if size > 0 {
		params["size"] = strconv.Itoa(size)
	}
	if offset > 0 {
		params["offset"] = strconv.Itoa(offset)
	}

	req, err := c.newRequest(ctx, "GET", "getAlbumList2", params)
	if err != nil {
		return nil, err
	}

	var resp Response
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	if err := c.checkError(resp.SubsonicResponse); err != nil {
		return nil, err
	}

	return resp.SubsonicResponse.AlbumList2, nil
}

func (c *Client) checkError(resp SubsonicResponse) error {
	if resp.Status == "failed" {
		if resp.Error != nil {
			return fmt.Errorf("api error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return fmt.Errorf("api returned failed status")
	}
	return nil
}

func (c *Client) GetStreamURL(id string) (string, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, "rest", "stream")

	q := u.Query()
	q.Set("id", id)
	q.Set("u", c.Username)
	q.Set("v", APIVersion)
	q.Set("c", ClientName)
	q.Set("f", "json")

	// Generate Auth Token
	salt, err := generateSalt(6)
	if err != nil {
		return "", fmt.Errorf("generate auth salt: %w", err)
	}
	token := md5Hash(c.Password + salt)
	q.Set("t", token)
	q.Set("s", salt)

	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (c *Client) newRequest(ctx context.Context, method, endpoint string, params map[string]string) (*http.Request, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "rest", endpoint)

	q := u.Query()
	q.Set("u", c.Username)
	q.Set("v", APIVersion)
	q.Set("c", ClientName)
	q.Set("f", "json") // Request JSON format

	// Generate Auth Token
	salt, err := generateSalt(6)
	if err != nil {
		return nil, fmt.Errorf("generate auth salt: %w", err)
	}
	token := md5Hash(c.Password + salt)
	q.Set("t", token)
	q.Set("s", salt)

	for k, v := range params {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()

	return http.NewRequestWithContext(ctx, method, u.String(), nil)
}

func (c *Client) do(req *http.Request, v interface{}) error {
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http error: %s", resp.Status)
	}

	// For debugging, you might want to print the body here
	// bodyBytes, _ := io.ReadAll(resp.Body)
	// fmt.Println(string(bodyBytes))
	// return json.Unmarshal(bodyBytes, v)

	return json.NewDecoder(resp.Body).Decode(v)
}

func generateRandomString(n int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rb := make([]byte, n)
	if _, err := cryptorand.Read(rb); err != nil {
		return "", err
	}

	out := make([]byte, n)
	for i := range out {
		out[i] = letters[int(rb[i])%len(letters)]
	}
	return string(out), nil
}

func md5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
