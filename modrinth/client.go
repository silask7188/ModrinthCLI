package modrinth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// @brief Client
// @base URL
// @http http Client
type Client struct {
	base *url.URL
	http *http.Client
}

// @brief new Client
// @param base url (always modrinth)
// @return new ready-to-use Client
func New(base string) (*Client, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	return &Client{
		base: u,
		http: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
			},
		},
	}, nil
}

// @brief JSON request
func (c *Client) doJSON(
	ctx context.Context,
	method string,
	path string,
	params url.Values,
	body io.Reader,
	dest any,
) error {

	u := c.base.ResolveReference(&url.URL{Path: path, RawQuery: params.Encode()})

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s %s: unexpected status %s", method, path, resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(dest)
}

// @brief GET json
func getJSON[T any](
	ctx context.Context,
	c *Client,
	path string,
	params url.Values,
) (*T, error) {
	var v T
	if err := c.doJSON(ctx, http.MethodGet, path, params, nil, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// @brief GET /search

func (c *Client) Search(ctx context.Context, p SearchParams) (*SearchResponse, error) {
	return getJSON[SearchResponse](ctx, c, "search", p.Values())
}

