// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2025 Steadybit GmbH

package extalert

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuery_StopsWhenPageReturnsNoEntries(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		// Report a total the server never fulfils (5 total, but no entries). Before the fix
		// this looped forever, hammering Splunk; now it must stop after one empty page.
		_, _ = w.Write([]byte(`{"paging":{"total":5},"entry":[]}`))
	}))
	defer srv.Close()

	c := &SplunkClient{client: resty.New().SetBaseURL(srv.URL)}

	entries, err := c.FiredAlerts(context.Background(), "/services/alerts")
	require.NoError(t, err)
	assert.Empty(t, entries)
	assert.Equal(t, 1, calls, "query must stop after an empty page instead of looping")
}
