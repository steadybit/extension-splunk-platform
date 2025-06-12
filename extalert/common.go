// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2025 Steadybit GmbH

package extalert

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/extension-splunk-platform/config"
	"strconv"
	"strings"
)

const (
	TargetType = "com.steadybit.extension_splunk_platform.alert"
	targetIcon = "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSI+PHBhdGggZmlsbC1ydWxlPSJldmVub2RkIiBjbGlwLXJ1bGU9ImV2ZW5vZGQiIGQ9Ik0xMiAyQzExLjE2MTQgMiAxMC40NDMzIDIuNTE2MTYgMTAuMTQ2MSAzLjI0ODEyQzcuMTc5ODMgNC4wNjA3MiA1IDYuNzc1NzkgNSAxMFYxNC42OTcyTDMuMTY3OTUgMTcuNDQ1M0MyLjk2MzM4IDE3Ljc1MjIgMi45NDQzMSAxOC4xNDY3IDMuMTE4MzMgMTguNDcxOUMzLjI5MjM1IDE4Ljc5NyAzLjYzMTIxIDE5IDQgMTlIOC41MzU0NEM4Ljc3ODA2IDIwLjY5NjEgMTAuMjM2OCAyMiAxMiAyMkMxMy43NjMyIDIyIDE1LjIyMTkgMjAuNjk2MSAxNS40NjQ2IDE5SDIwQzIwLjM2ODggMTkgMjAuNzA3NyAxOC43OTcgMjAuODgxNyAxOC40NzE5QzIxLjA1NTcgMTguMTQ2NyAyMS4wMzY2IDE3Ljc1MjIgMjAuODMyIDE3LjQ0NTNMMTkgMTQuNjk3MlYxMEMxOSA2Ljc3NTc5IDE2LjgyMDIgNC4wNjA3MiAxMy44NTM5IDMuMjQ4MTJDMTMuNTU2NyAyLjUxNjE2IDEyLjgzODYgMiAxMiAyWk0xMiAyMEMxMS4zNDY5IDIwIDEwLjc5MTMgMTkuNTgyNiAxMC41ODU0IDE5SDEzLjQxNDZDMTMuMjA4NyAxOS41ODI2IDEyLjY1MzEgMjAgMTIgMjBaTTE3IDEyLjQ1NDlWMTAuNTg0Nkw4IDZWOC4wNTI5NEwxNC45NzU0IDExLjVMOCAxNC45OTE1VjE3TDE3IDEyLjQ1OThWMTIuNDU0OVoiIGZpbGw9ImN1cnJlbnRDb2xvciIvPjwvc3ZnPg=="

	attributeID       = "splunk.alert.id"
	attributeName     = "splunk.alert.name"
	attributeAuthor   = "splunk.alert.author"
	attributeSeverity = "splunk.alert.severity"
	attributeUrl      = "splunk.alert.url"

	metricId          = "splunk.alert.metric.id"
	metricLabel       = "splunk.alert.metric.label"
	metricState       = "splunk.alert.metric.severity"
	metricTooltip     = "splunk.alert.metric.tooltip"
	metricTriggerTime = "splunk.alert.metric.triggerTime"
)

type SplunkClient struct {
	client *resty.Client
}

func NewSplunkClient() *SplunkClient {
	client := resty.New()
	if config.Config.DisableCertificateValidation {
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}) //NOSONAR explicit choice
	}
	client.SetBaseURL(strings.TrimRight(config.Config.ApiBaseUrl, "/"))
	client.SetHeader("Authorization", "Bearer "+config.Config.AccessToken)
	client.SetHeader("Content-Type", "application/json")
	return &SplunkClient{
		client: client,
	}
}

func (c *SplunkClient) Alerts(ctx context.Context) ([]Entry, error) {
	return c.query(ctx, "/services/saved/searches", map[string]string{
		"search": "alert.track=1",
	})
}

func (c *SplunkClient) FiredAlerts(ctx context.Context, alertUrl string) ([]Entry, error) {
	return c.query(ctx, alertUrl, nil)
}

func (c *SplunkClient) query(ctx context.Context, url string, params map[string]string) ([]Entry, error) {
	var entries []Entry
	total := 1

	for len(entries) < total {
		var response Response
		request := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParam("count", "30").
			SetQueryParam("offset", strconv.Itoa(len(entries))).
			SetQueryParam("output_mode", "json")

		if len(params) > 0 {
			for k, v := range params {
				request.SetQueryParam(k, v)
			}
		}

		res, err := request.Get(url)

		if err != nil {
			return nil, fmt.Errorf("failed to retrieve alerts from Splunk: %w", err)
		}

		if res.StatusCode() != 200 {
			return nil, fmt.Errorf("unexpected status code %d. full response: %v", res.StatusCode(), res.String())
		}

		log.Trace().Msgf("Splunk response (offset: %d): %v", len(entries), response)
		total = response.Paging.Total
		entries = append(entries, response.Entries...)
	}
	return entries, nil
}
