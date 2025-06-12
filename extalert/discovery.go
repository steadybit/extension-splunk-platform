// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2025 Steadybit GmbH

package extalert

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_commons"
	"github.com/steadybit/discovery-kit/go/discovery_kit_sdk"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-splunk-platform/config"
	"time"
)

type AlertClient interface {
	Alerts(ctx context.Context) ([]Entry, error)
}

type alertDiscovery struct {
	Client AlertClient
}

var (
	_ discovery_kit_sdk.TargetDescriber    = (*alertDiscovery)(nil)
	_ discovery_kit_sdk.AttributeDescriber = (*alertDiscovery)(nil)
)

func NewAlertDiscovery(client AlertClient) discovery_kit_sdk.TargetDiscovery {
	discovery := newAlertDiscovery(client)
	return discovery_kit_sdk.NewCachedTargetDiscovery(discovery,
		discovery_kit_sdk.WithRefreshTargetsNow(),
		discovery_kit_sdk.WithRefreshTargetsInterval(context.Background(), 1*time.Minute),
	)
}

func newAlertDiscovery(client AlertClient) *alertDiscovery {
	discovery := &alertDiscovery{
		Client: client,
	}
	return discovery
}

func (d *alertDiscovery) Describe() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id: TargetType,
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			CallInterval: extutil.Ptr("1m"),
		},
	}
}

func (d *alertDiscovery) DescribeTarget() discovery_kit_api.TargetDescription {
	return discovery_kit_api.TargetDescription{
		Id:       TargetType,
		Label:    discovery_kit_api.PluralLabel{One: "Alert", Other: "Alerts"},
		Category: extutil.Ptr("monitoring"),
		Version:  extbuild.GetSemverVersionStringOrUnknown(),
		Icon:     extutil.Ptr(targetIcon),
		Table: discovery_kit_api.Table{
			Columns: []discovery_kit_api.Column{
				{Attribute: attributeName},
			},
			OrderBy: []discovery_kit_api.OrderBy{
				{
					Attribute: attributeName,
					Direction: "ASC",
				},
			},
		},
	}
}

func (d *alertDiscovery) DescribeAttributes() []discovery_kit_api.AttributeDescription {
	return []discovery_kit_api.AttributeDescription{
		{
			Attribute: attributeID,
			Label: discovery_kit_api.PluralLabel{
				One:   "ID",
				Other: "IDs",
			},
		},
		{
			Attribute: attributeName,
			Label: discovery_kit_api.PluralLabel{
				One:   "Name",
				Other: "Names",
			},
		},
		{
			Attribute: attributeAuthor,
			Label: discovery_kit_api.PluralLabel{
				One:   "Author",
				Other: "Authors",
			},
		},
		{
			Attribute: attributeSeverity,
			Label: discovery_kit_api.PluralLabel{
				One:   "Severity",
				Other: "Severities",
			},
		},
		{
			Attribute: attributeUrl,
			Label: discovery_kit_api.PluralLabel{
				One:   "Fired Alert Url",
				Other: "Fired Alert Urls",
			},
		},
	}
}

func (d *alertDiscovery) DiscoverTargets(ctx context.Context) ([]discovery_kit_api.Target, error) {
	return d.getAllAlertTargets(ctx)
}

func (d *alertDiscovery) getAllAlertTargets(ctx context.Context) ([]discovery_kit_api.Target, error) {
	alerts, err := d.Client.Alerts(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to retrieve alerts")
		return make([]discovery_kit_api.Target, 0), err
	}

	result := make([]discovery_kit_api.Target, 0, len(alerts))
	for _, alert := range alerts {
		result = append(result, discovery_kit_api.Target{
			Id:         alert.Id,
			TargetType: TargetType,
			Label:      alert.Name,
			Attributes: map[string][]string{
				attributeID:       {alert.Id},
				attributeName:     {alert.Name},
				attributeAuthor:   {alert.Author},
				attributeSeverity: {alert.Content.Severity.String()},
				attributeUrl:      {alert.Links.Alerts},
			}})
	}
	return discovery_kit_commons.ApplyAttributeExcludes(result, config.Config.DiscoveryAttributesExcludesAlert), nil
}
