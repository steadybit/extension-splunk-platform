// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2025 Steadybit GmbH

package extalert

import (
	"context"
	"fmt"
	"github.com/steadybit/extension-splunk-platform/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAlertDiscovery_DiscoverTargets_noResponse(t *testing.T) {
	discovery := newAlertDiscovery(MockSplunkClient{
		response: []Entry{},
	})

	targets, err := discovery.getAllAlertTargets(context.Background())

	require.NoError(t, err)
	require.Empty(t, targets)
}

func TestAlertDiscovery_DiscoverTargets_multipleAlerts(t *testing.T) {
	discovery := newAlertDiscovery(MockSplunkClient{
		response: []Entry{
			{Id: "alert1", Name: "Alert One", Author: "Author1", Content: Content{Severity: SeverityFatal}},
			{Id: "alert2", Name: "Alert Two", Author: "Author2", Content: Content{Severity: SeverityDebug}},
		},
	})

	targets, err := discovery.getAllAlertTargets(context.Background())

	require.NoError(t, err)
	require.Len(t, targets, 2)

	require.Equal(t, "alert1", targets[0].Id)
	require.Equal(t, "Alert One", targets[0].Label)
	require.Equal(t, []string{"Fatal"}, targets[0].Attributes[attributeSeverity])
	require.Equal(t, []string{"Author1"}, targets[0].Attributes[attributeAuthor])

	require.Equal(t, "alert2", targets[1].Id)
	require.Equal(t, "Alert Two", targets[1].Label)
	require.Equal(t, []string{"Debug"}, targets[1].Attributes[attributeSeverity])
	require.Equal(t, []string{"Author2"}, targets[1].Attributes[attributeAuthor])
}

func TestAlertDiscovery_DiscoverTargets_invalidSeverity(t *testing.T) {
	discovery := newAlertDiscovery(MockSplunkClient{
		response: []Entry{
			{Id: "alert1", Name: "Alert One", Author: "Author1", Content: Content{Severity: 99}},
		},
	})

	targets, err := discovery.getAllAlertTargets(context.Background())

	require.NoError(t, err)
	require.Len(t, targets, 1)
	require.Equal(t, []string{"Unknown"}, targets[0].Attributes[attributeSeverity])
}

func TestAlertDiscovery_DiscoverTargets_excludedAttributes(t *testing.T) {
	config.Config.DiscoveryAttributesExcludesAlert = []string{attributeAuthor}
	defer func() {
		config.Config.DiscoveryAttributesExcludesAlert = []string{}
	}()

	discovery := newAlertDiscovery(MockSplunkClient{
		response: []Entry{
			{Id: "alert1", Name: "Alert One", Author: "Author1", Content: Content{Severity: SeverityInfo}},
		},
	})

	targets, err := discovery.getAllAlertTargets(context.Background())

	require.NoError(t, err)
	require.Len(t, targets, 1)
	require.NotContains(t, targets[0].Attributes, attributeAuthor)
}

func TestAlertDiscovery_DiscoverTargets_errorResponse(t *testing.T) {
	discovery := newAlertDiscovery(MockSplunkClient{
		err: fmt.Errorf("some error"),
	})

	targets, err := discovery.getAllAlertTargets(context.Background())

	require.Empty(t, targets)
	require.Error(t, err)
}
