package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/penkovgd/closer"
	"yadro.com/course/update/core"
)

type Client struct {
	log    *slog.Logger
	client http.Client
	url    string
}

func NewClient(url string, timeout time.Duration, log *slog.Logger) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("empty base url specified")
	}
	return &Client{
		client: http.Client{Timeout: timeout},
		log:    log,
		url:    url,
	}, nil
}

type GetResponse struct {
	Num        int    `json:"num"`
	Transcript string `json:"transcript"`
	Title      string `json:"title"`
	SafeTitle  string `json:"safe_title"`
	Alt        string `json:"alt"`
	Img        string `json:"img"`
	Day        string `json:"day"`
	Month      string `json:"month"`
	Year       string `json:"year"`
}

func (c Client) Get(ctx context.Context, id int) (core.XKCDInfo, error) {
	endpoint := fmt.Sprintf(`%s/%d/info.0.json`, c.url, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		c.log.Error("failed to create request", "endpoint", endpoint, "error", err)
		return core.XKCDInfo{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error("failed to fetch comic", "id", id, "error", err)
		return core.XKCDInfo{}, err
	}
	defer closer.CloseOrLog(c.log, resp.Body)

	if resp.StatusCode != http.StatusOK {

		if resp.StatusCode == http.StatusNotFound {
			c.log.Error("comic not found", "status", resp.StatusCode, "id", id)
			return core.XKCDInfo{}, core.ErrNotFound
		}

		c.log.Error("received non-OK http status", "status", resp.StatusCode, "id", id)
		return core.XKCDInfo{}, fmt.Errorf("received non-OK http status: %d", resp.StatusCode)
	}

	var respDecoded GetResponse
	if err := json.NewDecoder(resp.Body).Decode(&respDecoded); err != nil {
		c.log.Error("failed to decode json response", "id", id, "error", err)
		return core.XKCDInfo{}, err
	}

	year, err := strconv.Atoi(respDecoded.Year)
	if err != nil {
		c.log.Error("failed to convert to int", "str", respDecoded.Year, "error", err)
		return core.XKCDInfo{}, err
	}
	month, err := strconv.Atoi(respDecoded.Month)
	if err != nil {
		c.log.Error("failed to convert to int", "str", respDecoded.Month, "error", err)
		return core.XKCDInfo{}, err
	}
	day, err := strconv.Atoi(respDecoded.Day)
	if err != nil {
		c.log.Error("failed to convert to int", "str", respDecoded.Day, "error", err)
		return core.XKCDInfo{}, err
	}

	c.log.Debug("successfully fetched comic", "id", id)

	return core.XKCDInfo{
		ID:          respDecoded.Num,
		URL:         respDecoded.Img,
		Title:       respDecoded.Title,
		SafeTitle:   respDecoded.SafeTitle,
		Description: respDecoded.Transcript,
		Alt:         respDecoded.Alt,
		Date: time.Date(
			year,
			time.Month(month),
			day,
			0, 0, 0, 0,
			time.UTC,
		),
	}, nil
}

type LastIDResponse struct {
	Id int `json:"num"`
}

func (c Client) LastID(ctx context.Context) (int, error) {
	endpoint := c.url + `/info.0.json`

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		c.log.Error("failed to create request", "endpoint", endpoint, "error", err)
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error("failed to send request ", "endpoint", endpoint, "error", err)
		return 0, err
	}
	defer closer.CloseOrLog(c.log, resp.Body)

	if resp.StatusCode != http.StatusOK {
		c.log.Error("received non-OK http status", "status", resp.StatusCode)
		return 0, fmt.Errorf("received non-OK http status: %d", resp.StatusCode)
	}

	var lastIDResp LastIDResponse
	if err := json.NewDecoder(resp.Body).Decode(&lastIDResp); err != nil {
		c.log.Error("failed to decode json response", "error", err)
		return 0, err
	}

	c.log.Debug("successfully fetched last id", "id", lastIDResp.Id)

	return lastIDResp.Id, nil
}

func (c Client) GetImage(ctx context.Context, imageURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer closer.CloseOrLog(c.log, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("fetch image: unexpected status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	ext := chooseExtension(imageURL, contentType)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return data, ext, nil
}

func chooseExtension(imageURL, contentType string) string {
	if contentType != "" {
		if exts, _ := mime.ExtensionsByType(contentType); len(exts) > 0 {
			return exts[0] // ".png", ".jpg"
		}
	}
	ext := path.Ext(strings.Split(imageURL, "?")[0])
	if ext == "" {
		return ".bin"
	}
	return ext
}
