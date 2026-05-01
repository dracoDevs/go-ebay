package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/dracoDevs/go-ebay/pkg/auth"
	"github.com/dracoDevs/go-ebay/pkg/internal/rest"
)

const baseURL = "https://apim.ebay.com/commerce/media/v1_beta"

type Client struct {
	doer rest.Doer
}

type Option func(*Client)

func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.doer.HTTPClient = c }
}

func WithBaseURL(u string) Option {
	return func(cl *Client) { cl.doer.BaseURL = u }
}

func NewClient(src auth.TokenSource, opts ...Option) *Client {
	c := &Client{
		doer: rest.Doer{
			TokenSource: src,
			HTTPClient:  &http.Client{Timeout: 30 * time.Second},
			BaseURL:     baseURL,
			ErrPrefix:   "media:",
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

type createImageFromURLRequest struct {
	ImageURL string `json:"imageUrl"`
}

// CreateImageFromURL ingests an image hosted at imageURL into eBay's
// Picture Hosting Service (EPS) and returns the eBay-hosted image URL
// the caller can pass into Inventory product.imageUrls. Replaces
// Trading UploadSiteHostedPicture for the URL-fetch case.
func (c *Client) CreateImageFromURL(ctx context.Context, imageURL string) (string, error) {
	if imageURL == "" {
		return "", fmt.Errorf("media: imageURL is required")
	}
	body, _ := json.Marshal(createImageFromURLRequest{ImageURL: imageURL})
	res, err := c.doer.Do(ctx, http.MethodPost, "/image/create_image_from_url", bytes.NewReader(body), "application/json")
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("media: createImageFromURL %d: %s", res.StatusCode, string(res.Body))
	}

	var bodyOut struct {
		ImageURL string `json:"imageUrl"`
		ImageID  string `json:"imageId"`
	}
	if err := json.Unmarshal(res.Body, &bodyOut); err == nil && bodyOut.ImageURL != "" {
		return bodyOut.ImageURL, nil
	}

	if loc := res.Header.Get("Location"); loc != "" {
		if u, perr := url.Parse(loc); perr == nil {
			id := path.Base(u.Path)
			if id != "" && id != "/" && id != "." {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("media: createImageFromURL: no imageUrl in response (body: %s)", string(res.Body))
}
