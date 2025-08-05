package testclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/PROJECT_NAME/internal/utils"
	"github.com/PROJECT_NAME/internal/utils/iolimit"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func init() {
	// this seems to work even if the init happens before setting up the trace provider
	tracer = otel.Tracer("github.com/PROJECT_NAME/internal/clients/testclient")
}

var _ Client = new(client)

type (
	Client interface {
		GetFancyResponse(ctx context.Context) (*FancyResponseFromTestClient, error)
		IsReady() error
	}

	ClientProvider interface {
		TestClient() Client
	}

	clientDependencies interface {
		config.ConfigProvider
		logger.LoggerProvider
		utils.RetryProvider
	}

	client struct {
		d clientDependencies
		c *http.Client
	}
)

func NewClient(d clientDependencies) *client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   d.Config().Clients.HttpTransport.DialTimeout,
			KeepAlive: d.Config().Clients.HttpTransport.DialKeepAlive,
		}).DialContext,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        d.Config().Clients.HttpTransport.MaxIdleConns,
		IdleConnTimeout:     d.Config().Clients.HttpTransport.IdleConnTimeout,
		TLSHandshakeTimeout: d.Config().Clients.HttpTransport.TLSHandshakeTimeout,
		DisableKeepAlives:   d.Config().Clients.HttpTransport.DisableKeepAlive,
	}

	return &client{
		d: d,
		c: &http.Client{
			Transport: transport,
			Timeout:   d.Config().Clients.Timeout,
		},
	}
}

func (c *client) GetFancyResponse(ctx context.Context) (*FancyResponseFromTestClient, error) {
	var resp FancyResponseFromTestClient
	if err := c.makeRequest(ctx, http.MethodGet, "/api/wow/something/nice", nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *client) IsReady() error {
	return c.makeRequest(context.Background(), http.MethodGet, "/api/ping", nil, nil)
}

func (c *client) makeRequest(ctx context.Context, method, path string, body, resp interface{}, headers ...http.Header) error {
	startTime := time.Now()

	var bodyBytes []byte
	var err error

	_, span := tracer.Start(ctx, "testclient-request")
	defer span.End()

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			span.RecordError(err)
			return err
		}
		c.d.Logger().Debugw("TestClient Request body for", "path", path, "body", string(bodyBytes))
		span.SetAttributes(attribute.String("request_body", string(bodyBytes)))
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s/%s", c.d.Config().TestClient.BaseUrl, utils.NormalizePath(path)), bytes.NewBuffer(bodyBytes))
	if err != nil {
		span.RecordError(err)
		return err
	}

	for _, header := range headers {
		req.Header = header
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.d.Config().TestClient.ApiKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Nice-Advice", "Stay curious and keep learning!")

	response, err := c.d.Retry().DoHttp("testclient-request", c.c, req)
	if err != nil {
		span.RecordError(err)
		return err
	}
	defer response.Body.Close()

	respBody, err := iolimit.ReadAll(response.Body, c.d.Config().GetClientMaxRegularBodySize())
	if err != nil {
		span.RecordError(err)
		return err
	}

	c.d.Logger().Debugw("TestClient Response body for", "path", path, "body", string(respBody))
	span.SetAttributes(attribute.String("response_body", string(respBody)))
	span.SetAttributes(attribute.String("response_time", fmt.Sprintf("%dms", time.Since(startTime).Milliseconds())))

	if response.StatusCode != http.StatusOK {
		c.d.Logger().Errorw("TestClient Request failed", "path", path, "status_code", response.StatusCode)

		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to unmarshal error response: %w", err)
		}

		return errResp.mapError()
	}

	if resp != nil {
		if err := json.Unmarshal(respBody, resp); err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
