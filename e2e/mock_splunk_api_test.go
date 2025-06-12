// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2025 Steadybit GmbH

package e2e

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-splunk-platform/extalert"
	"net"
	"net/http"
	"net/http/httptest"
)

type mockServer struct {
	http *httptest.Server
}

func createMockSplunkApiServer() *mockServer {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		panic(fmt.Sprintf("httptest: failed to listen: %v", err))
	}
	mux := http.NewServeMux()

	server := httptest.Server{Listener: listener, Config: &http.Server{Handler: mux}}
	server.Start()
	log.Info().Str("url", server.URL).Msg("Started Mock-Server")

	mock := &mockServer{http: &server}
	mux.Handle("GET /services/saved/searches", handler(mock.getSavedSearches))
	mux.Handle("GET /servicesNS/nobody/myTestApp/user/alerts/Enty%201", handler(mock.getFiredAlerts))
	return mock
}

func handler[T any](getter func() T) http.Handler {
	return exthttp.PanicRecovery(exthttp.LogRequest(exthttp.GetterAsHandler(getter)))
}

func (m *mockServer) getSavedSearches() extalert.Response {
	return extalert.Response{
		Paging: extalert.Paging{
			Total:   2,
			PerPage: 30,
			Offset:  0,
		},
		Entries: []extalert.Entry{
			{
				Id:     "entry-1",
				Name:   "Entry 1",
				Author: "e2e test",
				Content: extalert.Content{
					Severity: extalert.SeveritySevere,
				},
				Links: extalert.Links{
					Alerts: "/servicesNS/nobody/myTestApp/user/alerts/Enty%201",
				},
			},
		},
	}
}

func (m *mockServer) getFiredAlerts() extalert.Response {
	return extalert.Response{
		Paging: extalert.Paging{
			Total:   1,
			PerPage: 30,
			Offset:  0,
		},
		Entries: []extalert.Entry{
			{
				Id:     "entry-1",
				Name:   "Entry 1",
				Author: "e2e test",
				Content: extalert.Content{
					Severity:    extalert.SeveritySevere,
					TriggerTime: 946684800,
				},
			},
		},
	}
}
