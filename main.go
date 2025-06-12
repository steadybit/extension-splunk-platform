// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2025 Steadybit GmbH

package main

import (
	_ "github.com/KimMachineGun/automemlimit" // By default, it sets `GOMEMLIMIT` to 90% of cgroup's memory limit.
	"github.com/rs/zerolog"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_sdk"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthealth"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-kit/extruntime"
	"github.com/steadybit/extension-kit/extsignals"
	"github.com/steadybit/extension-splunk-platform/config"
	"github.com/steadybit/extension-splunk-platform/extalert"
	_ "go.uber.org/automaxprocs" // Importing automaxprocs automatically adjusts GOMAXPROCS.
)

func main() {
	extlogging.InitZeroLog()

	extbuild.PrintBuildInformation()
	extruntime.LogRuntimeInformation(zerolog.DebugLevel)

	exthealth.SetReady(false)
	exthealth.StartProbes(8084)

	config.ParseConfiguration()

	exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(getExtensionList))

	splunkClient := extalert.NewSplunkClient()
	discovery_kit_sdk.Register(extalert.NewAlertDiscovery(splunkClient))
	action_kit_sdk.RegisterAction(extalert.NewAlertCheckAction(splunkClient))

	extsignals.ActivateSignalHandlers()
	action_kit_sdk.RegisterCoverageEndpoints()
	exthealth.SetReady(true)

	exthttp.Listen(exthttp.ListenOpts{
		Port: 8083,
	})
}

type ExtensionListResponse struct {
	action_kit_api.ActionList       `json:",inline"`
	discovery_kit_api.DiscoveryList `json:",inline"`
}

func getExtensionList() ExtensionListResponse {
	return ExtensionListResponse{
		ActionList:    action_kit_sdk.GetActionList(),
		DiscoveryList: discovery_kit_sdk.GetDiscoveryList(),
	}
}
