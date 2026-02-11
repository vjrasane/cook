package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.listonic.com"
	loginEndpoint  = "/api/loginextended"
	listsEndpoint  = "/api/lists"

	clientID     = "listonicv2"
	clientSecret = "fjdfsoj9874jdfhjkh34jkhffdfff"
	redirectURI  = "https://listonicv2api.jestemkucharzem.pl"
)

var clientAuth = base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	token      string
	email      string
	password   string
}

func NewClient(email, password string) *Client {
	return &Client{
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{},
		email:      email,
		password:   password,
	}
}

func (c *Client) Authenticate() error {
	cached, _ := loadTokenCache()
	if cached != nil && cached.Valid() {
		c.token = cached.AccessToken
		return nil
	}

	if cached != nil && cached.RefreshToken != "" {
		if err := c.refresh(cached.RefreshToken); err == nil {
			return nil
		}
	}

	return c.login()
}

func (c *Client) login() error {
	if c.email == "" || c.password == "" {
		return fmt.Errorf("LISTONIC_EMAIL and LISTONIC_PASSWORD must be set")
	}

	params := url.Values{}
	params.Set("provider", "password")
	params.Set("autoMerge", "1")
	params.Set("autoDestruct", "1")

	form := url.Values{}
	form.Set("username", c.email)
	form.Set("password", c.password)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("redirect_uri", redirectURI)

	authResp, err := c.postAuth(params, form)
	if err != nil {
		return err
	}

	c.token = authResp.AccessToken
	saveTokenCache(&TokenCache{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    time.Now().Unix() + int64(authResp.ExpiresIn),
	})
	return nil
}

func (c *Client) refresh(refreshToken string) error {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	authResp, err := c.postAuth(url.Values{}, form)
	if err != nil {
		return err
	}

	c.token = authResp.AccessToken
	saveTokenCache(&TokenCache{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresAt:    time.Now().Unix() + int64(authResp.ExpiresIn),
	})
	return nil
}

func (c *Client) postAuth(params, form url.Values) (*AuthResponse, error) {
	reqURL := c.BaseURL + loginEndpoint
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}
	req, err := http.NewRequest("POST", reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("clientauthorization", "Bearer "+clientAuth)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading auth response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("authentication failed (%d): %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, fmt.Errorf("parsing auth response: %w", err)
	}

	if authResp.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response")
	}

	return &authResp, nil
}

func (c *Client) doJSON(method, path string, reqBody any) ([]byte, int, error) {
	var bodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return nil, 0, fmt.Errorf("marshaling request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response: %w", err)
	}

	return body, resp.StatusCode, nil
}

func (c *Client) GetLists() ([]List, error) {
	body, status, err := c.doJSON("GET", listsEndpoint+"?includeShares=true&archive=false&includeItems=false", nil)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, fmt.Errorf("get lists failed (%d): %s", status, string(body))
	}

	var lists []List
	if err := json.Unmarshal(body, &lists); err != nil {
		return nil, fmt.Errorf("parsing lists: %w", err)
	}
	return lists, nil
}

func (c *Client) GetList(listID string) (*List, error) {
	body, status, err := c.doJSON("GET", listsEndpoint+"/"+listID, nil)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, fmt.Errorf("get list failed (%d): %s", status, string(body))
	}

	var list List
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("parsing list: %w", err)
	}
	return &list, nil
}

func (c *Client) GetListItems(listID string) ([]Item, error) {
	body, status, err := c.doJSON("GET", listsEndpoint+"/"+listID+"/items", nil)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, fmt.Errorf("get items failed (%d): %s", status, string(body))
	}

	var items []Item
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("parsing items: %w", err)
	}
	return items, nil
}

func (c *Client) AddItem(listID string, item AddItemRequest) (*Item, error) {
	body, status, err := c.doJSON("POST", listsEndpoint+"/"+listID+"/items", item)
	if err != nil {
		return nil, err
	}
	if status != 200 && status != 201 {
		return nil, fmt.Errorf("add item failed (%d): %s", status, string(body))
	}

	var created Item
	if err := json.Unmarshal(body, &created); err != nil {
		return nil, fmt.Errorf("parsing item: %w", err)
	}
	return &created, nil
}

func (c *Client) UpdateItem(listID, itemID string, update UpdateItemRequest) error {
	body, status, err := c.doJSON("PATCH", listsEndpoint+"/"+listID+"/items/"+itemID, update)
	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("update item failed (%d): %s", status, string(body))
	}
	return nil
}

func (c *Client) DeleteItem(listID, itemID string) error {
	body, status, err := c.doJSON("DELETE", listsEndpoint+"/"+listID+"/items/"+itemID, nil)
	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("delete item failed (%d): %s", status, string(body))
	}
	return nil
}

func (c *Client) ClearItems(listID string, checkedOnly bool) ([]string, error) {
	items, err := c.GetListItems(listID)
	if err != nil {
		return nil, err
	}

	var deleted []string
	for _, item := range items {
		if checkedOnly && item.Checked == 0 {
			continue
		}
		if err := c.DeleteItem(listID, item.Id); err != nil {
			return deleted, fmt.Errorf("deleting item %s: %w", item.Id, err)
		}
		deleted = append(deleted, item.Id)
	}
	return deleted, nil
}

func (c *Client) CreateList(name string) (*List, error) {
	body, status, err := c.doJSON("POST", listsEndpoint, CreateListRequest{Name: name})
	if err != nil {
		return nil, err
	}
	if status != 200 && status != 201 {
		return nil, fmt.Errorf("create list failed (%d): %s", status, string(body))
	}

	var created List
	if err := json.Unmarshal(body, &created); err != nil {
		return nil, fmt.Errorf("parsing list: %w", err)
	}
	return &created, nil
}

func (c *Client) DeleteList(listID string) error {
	body, status, err := c.doJSON("DELETE", listsEndpoint+"/"+listID, nil)
	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("delete list failed (%d): %s", status, string(body))
	}
	return nil
}

func (c *Client) UpdateList(listID string, name string) error {
	body, status, err := c.doJSON("PATCH", listsEndpoint+"/"+listID, UpdateListRequest{Name: name})
	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("update list failed (%d): %s", status, string(body))
	}
	return nil
}

func (c *Client) ResolveListID(nameOrID string) (string, error) {
	if _, err := strconv.Atoi(nameOrID); err == nil {
		return nameOrID, nil
	}

	lists, err := c.GetLists()
	if err != nil {
		return "", fmt.Errorf("fetching lists for name lookup: %w", err)
	}

	lower := strings.ToLower(nameOrID)
	for _, l := range lists {
		if strings.ToLower(l.Name) == lower {
			return l.Id, nil
		}
	}

	return "", fmt.Errorf("list not found: %q", nameOrID)
}
