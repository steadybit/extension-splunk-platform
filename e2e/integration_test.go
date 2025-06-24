// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2025 Steadybit GmbH

package e2e

import (
	"context"
	"fmt"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_test/e2e"
	actValidate "github.com/steadybit/action-kit/go/action_kit_test/validate"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_test/validate"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestWithMinikube(t *testing.T) {
	extlogging.InitZeroLog()

	// Generate self-signed certificate for testing
	cleanup, err := generateSelfSignedCert()
	require.NoError(t, err)
	defer cleanup()

	server := createMockSplunkApiServer()
	defer server.http.Close()
	split := strings.SplitAfter(server.http.URL, ":")
	port := split[len(split)-1]

	// Test with insecureSkipVerify approach
	t.Run("with insecureSkipVerify", func(t *testing.T) {
		extFactory := e2e.HelmExtensionFactory{
			Name: "extension-splunk-platform",
			Port: 8083,
			ExtraArgs: func(m *e2e.Minikube) []string {
				return []string{
					"--set", fmt.Sprintf("splunk.apiBaseUrl=https://host.minikube.internal:%s", port),
					"--set", "logging.level=trace",
					"--set", "splunk.insecureSkipVerify=true", // Enable skipping TLS verification
				}
			},
		}

		e2e.WithDefaultMinikube(t, &extFactory, []e2e.WithMinikubeTestCase{
			{
				Name: "test discovery with insecureSkipVerify",
				Test: testDiscovery,
			},
			{
				Name: "validate discovery with insecureSkipVerify",
				Test: validateDiscovery,
			},
			{
				Name: "validate actions with insecureSkipVerify",
				Test: validateActions,
			},
			{
				Name: "test check action with insecureSkipVerify",
				Test: testCheckAction,
			},
		})
	})

	// Test with custom certificate approach
	t.Run("with custom certificate", func(t *testing.T) {
		extFactory := e2e.HelmExtensionFactory{
			Name: "extension-splunk-platform",
			Port: 8083,
			ExtraArgs: func(m *e2e.Minikube) []string {
				return []string{
					"--set", fmt.Sprintf("splunk.apiBaseUrl=https://host.minikube.internal:%s", port),
					"--set", "logging.level=trace",
					"--set", "splunk.insecureSkipVerify=false", // Disable insecureSkipVerify
					// Use extraVolumeMounts, extraVolumes and extraEnv instead
					"--set", "extraVolumeMounts[0].name=extra-certs",
					"--set", "extraVolumeMounts[0].mountPath=/etc/ssl/extra-certs",
					"--set", "extraVolumeMounts[0].readOnly=true",
					"--set", "extraVolumes[0].name=extra-certs",
					"--set", "extraVolumes[0].configMap.name=splunk-self-signed-ca",
					"--set", "extraEnv[0].name=SSL_CERT_DIR",
					"--set", "extraEnv[0].value=/etc/ssl/extra-certs:/etc/ssl/certs",
				}
			},
		}

		e2e.WithMinikube(t, e2e.DefaultMinikubeOpts().AfterStart(installConfigMap), &extFactory, []e2e.WithMinikubeTestCase{
			{
				Name: "test discovery with custom certificate",
				Test: testDiscovery,
			},
			{
				Name: "validate discovery with custom certificate",
				Test: validateDiscovery,
			},
			{
				Name: "validate actions with custom certificate",
				Test: validateActions,
			},
			{
				Name: "test check action with custom certificate",
				Test: testCheckAction,
			},
		})
	})
}

func validateDiscovery(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	assert.NoError(t, validate.ValidateEndpointReferences("/", e.Client))
}

func testDiscovery(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	target, err := e2e.PollForTarget(ctx, e, "com.steadybit.extension_splunk_platform.alert", func(target discovery_kit_api.Target) bool {
		return e2e.HasAttribute(target, "splunk.alert.id", "entry-1")
	})
	require.NoError(t, err)
	assert.Equal(t, target.TargetType, "com.steadybit.extension_splunk_platform.alert")
	assert.Equal(t, target.Attributes["splunk.alert.id"], []string{"entry-1"})
	assert.Equal(t, target.Attributes["splunk.alert.name"], []string{"Entry 1"})
	assert.Equal(t, target.Attributes["splunk.alert.author"], []string{"e2e test"})
	assert.Equal(t, target.Attributes["splunk.alert.severity"], []string{"Severe"})
	assert.Equal(t, target.Attributes["splunk.alert.url"], []string{"/servicesNS/nobody/myTestApp/user/alerts/Enty%201"})
}

func validateActions(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	assert.NoError(t, actValidate.ValidateEndpointReferences("/", e.Client))
}

func testCheckAction(t *testing.T, minikube *e2e.Minikube, e *e2e.Extension) {
	target := &action_kit_api.Target{
		Name: "test_check",
		Attributes: map[string][]string{
			"splunk.alert.id":       {"entry-1"},
			"splunk.alert.name":     {"Entry 1"},
			"splunk.alert.author":   {"e2e test"},
			"splunk.alert.severity": {"Severe"},
			"splunk.alert.url":      {"/servicesNS/nobody/myTestApp/user/alerts/Enty%201"},
		},
	}

	config := struct {
		Duration           int    `json:"duration"`
		CheckNewAlertsOnly string `json:"checkNewAlertsOnly"`
		ExpectedState      string `json:"expectedState"`
		StateCheckMode     string `json:"stateCheckMode"`
	}{
		Duration:           1_000,
		CheckNewAlertsOnly: "false",
		ExpectedState:      "alertFired",
		StateCheckMode:     "atLeastOnce",
	}

	action, err := e.RunAction("com.steadybit.extension_splunk_platform.alert.check", target, config, &action_kit_api.ExecutionContext{})
	require.NoError(t, err)
	defer func() { _ = action.Cancel() }()

	require.NoError(t, action.Wait())
	require.NotEmpty(t, t, action.Messages())
	require.NotEmpty(t, t, action.Metrics())
}
