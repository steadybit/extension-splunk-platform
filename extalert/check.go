// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2025 Steadybit GmbH

package extalert

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
	"time"
)

type FiredAlertsClient interface {
	FiredAlerts(ctx context.Context, alertUrl string) ([]Entry, error)
}

type AlertCheckAction struct {
	Client FiredAlertsClient
}

var (
	_ action_kit_sdk.Action[AlertCheckState]           = (*AlertCheckAction)(nil)
	_ action_kit_sdk.ActionWithStatus[AlertCheckState] = (*AlertCheckAction)(nil)
)

type AlertCheckState struct {
	Id                 string
	Name               string
	Url                string
	Start              time.Time
	End                time.Time
	CheckNewAlertsOnly bool
	ExpectedState      string
	StateCheckMode     string
	StateCheckSuccess  bool
	TriggerTime        int64
}

const (
	alertFired    = "alertFired"
	alertNotFired = "alertNotFired"

	stateCheckModeAtLeastOnce = "atLeastOnce"
	stateCheckModeAllTheTime  = "allTheTime"
)

func NewAlertCheckAction(client FiredAlertsClient) action_kit_sdk.Action[AlertCheckState] {
	return &AlertCheckAction{
		Client: client,
	}
}

func (a *AlertCheckAction) NewEmptyState() AlertCheckState {
	return AlertCheckState{}
}

func (a *AlertCheckAction) Describe() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          fmt.Sprintf("%s.check", TargetType),
		Label:       "Alert Status",
		Description: "Check the status of an alert.",
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Icon:        extutil.Ptr(targetIcon),
		TargetSelection: extutil.Ptr(action_kit_api.TargetSelection{
			TargetType:          TargetType,
			QuantityRestriction: extutil.Ptr(action_kit_api.All),
			SelectionTemplates: extutil.Ptr([]action_kit_api.TargetSelectionTemplate{
				{
					Label:       "Alert name",
					Description: extutil.Ptr("Find alert by name"),
					Query:       attributeName + "=\"\"",
				},
			}),
		}),
		Technology:  extutil.Ptr("Splunk"),
		Category:    extutil.Ptr("Monitoring"),
		Kind:        action_kit_api.Check,
		TimeControl: action_kit_api.TimeControlInternal,
		Parameters: []action_kit_api.ActionParameter{
			{
				Name:         "duration",
				Label:        "Duration",
				Type:         action_kit_api.ActionParameterTypeDuration,
				DefaultValue: extutil.Ptr("30s"),
				Required:     extutil.Ptr(true),
			},
			{
				Name:         "checkNewAlertsOnly",
				Label:        "New Alerts Only",
				Description:  extutil.Ptr("Only check events fired after the start of the experiment."),
				Type:         action_kit_api.ActionParameterTypeBoolean,
				DefaultValue: extutil.Ptr("false"),
				Required:     extutil.Ptr(true),
			},
			{
				Name:         "expectedState",
				Label:        "Expected State",
				Type:         action_kit_api.ActionParameterTypeString,
				DefaultValue: extutil.Ptr(alertFired),
				Options: extutil.Ptr([]action_kit_api.ParameterOption{
					action_kit_api.ExplicitParameterOption{
						Label: "Alert fired",
						Value: alertFired,
					},
					action_kit_api.ExplicitParameterOption{
						Label: "Alert not fired",
						Value: alertNotFired,
					},
				}),
				Required: extutil.Ptr(true),
			},
			{
				Name:         "stateCheckMode",
				Label:        "State Check Mode",
				Description:  extutil.Ptr("How often should the observed state match the expectation?"),
				Type:         action_kit_api.ActionParameterTypeString,
				DefaultValue: extutil.Ptr(stateCheckModeAtLeastOnce),
				Options: extutil.Ptr([]action_kit_api.ParameterOption{
					action_kit_api.ExplicitParameterOption{
						Label: "All the time",
						Value: stateCheckModeAllTheTime,
					},
					action_kit_api.ExplicitParameterOption{
						Label: "At least once",
						Value: stateCheckModeAtLeastOnce,
					},
				}),
				Required: extutil.Ptr(true),
			},
		},
		Widgets: extutil.Ptr([]action_kit_api.Widget{
			action_kit_api.StateOverTimeWidget{
				Type:  action_kit_api.ComSteadybitWidgetStateOverTime,
				Title: "Alert State",
				Identity: action_kit_api.StateOverTimeWidgetIdentityConfig{
					From: metricId,
				},
				Label: action_kit_api.StateOverTimeWidgetLabelConfig{
					From: metricLabel,
				},
				State: action_kit_api.StateOverTimeWidgetStateConfig{
					From: metricState,
				},
				Tooltip: action_kit_api.StateOverTimeWidgetTooltipConfig{
					From: metricTooltip,
				},
				Value: extutil.Ptr(action_kit_api.StateOverTimeWidgetValueConfig{
					Hide: extutil.Ptr(true),
				}),
			},
		}),
		Status: extutil.Ptr(action_kit_api.MutatingEndpointReferenceWithCallInterval{
			CallInterval: extutil.Ptr("1s"),
		}),
	}
}

func (a *AlertCheckAction) Prepare(_ context.Context, state *AlertCheckState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	alertId := request.Target.Attributes[attributeID]
	if len(alertId) == 0 {
		return nil, fmt.Errorf("target is missing the id attribute")
	}

	start := time.Now()
	end := start.Add(time.Duration(extutil.ToInt64(request.Config["duration"])) * time.Millisecond)

	expectedState := extutil.ToString(request.Config["expectedState"])
	if expectedState == "" {
		return nil, fmt.Errorf("expected state parameter is missing")
	}

	stateCheckMode := extutil.ToString(request.Config["stateCheckMode"])
	if stateCheckMode == "" {
		return nil, fmt.Errorf("expected state check mode parameter is missing")
	}

	checkNewAlertsOnly := extutil.ToBool(request.Config["checkNewAlertsOnly"])

	state.Id = alertId[0]
	state.Name = request.Target.Attributes[attributeName][0]
	state.Url = request.Target.Attributes[attributeUrl][0]
	state.CheckNewAlertsOnly = checkNewAlertsOnly
	state.Start = start
	state.End = end
	state.ExpectedState = expectedState
	state.StateCheckMode = stateCheckMode

	log.Trace().Any("state", state).Msg("check action state")

	return nil, nil
}

func (a *AlertCheckAction) Start(_ context.Context, _ *AlertCheckState) (*action_kit_api.StartResult, error) {
	return nil, nil
}

func (a *AlertCheckAction) Status(ctx context.Context, state *AlertCheckState) (*action_kit_api.StatusResult, error) {
	return checkFiredAlerts(ctx, state, a.Client)
}

func checkFiredAlerts(ctx context.Context, state *AlertCheckState, client FiredAlertsClient) (*action_kit_api.StatusResult, error) {
	now := time.Now()

	allFiredAlerts, err := client.FiredAlerts(ctx, state.Url)
	if err != nil {
		return nil, err
	}

	var firedAlerts []Entry
	if state.CheckNewAlertsOnly {
		for _, firedAlert := range allFiredAlerts {
			if firedAlert.Content.TriggerTime > state.Start.Unix() {
				firedAlerts = append(firedAlerts, firedAlert)
			}
		}
	} else {
		firedAlerts = allFiredAlerts
	}

	completed := now.After(state.End)
	var checkError *action_kit_api.ActionKitError
	if state.StateCheckMode == stateCheckModeAllTheTime {
		checkError = checkAllTheTime(state, firedAlerts)
	} else if state.StateCheckMode == stateCheckModeAtLeastOnce {
		checkError = checkAtLeastOnce(state, completed, firedAlerts)
	}

	return &action_kit_api.StatusResult{
		Completed: completed,
		Error:     checkError,
		Metrics:   extutil.Ptr(toMetrics(state.Name, firedAlerts, now)),
	}, nil
}

func checkAllTheTime(state *AlertCheckState, firedAlerts []Entry) *action_kit_api.ActionKitError {
	if state.ExpectedState == alertNotFired && len(firedAlerts) > 0 {
		triggerTime := time.Unix(firedAlerts[0].Content.TriggerTime, 0).UTC().Format(time.RFC3339)
		return extutil.Ptr(action_kit_api.ActionKitError{
			Title:  fmt.Sprintf("Alert %q should not have been fired but was at %s.", state.Name, triggerTime),
			Status: extutil.Ptr(action_kit_api.Failed),
		})
	}

	if state.ExpectedState == alertFired && len(firedAlerts) == 0 {
		return extutil.Ptr(action_kit_api.ActionKitError{
			Title:  fmt.Sprintf("Alert %q should have been fired all the time but was not.", state.Name),
			Status: extutil.Ptr(action_kit_api.Failed),
		})
	}

	return nil
}

func checkAtLeastOnce(state *AlertCheckState, completed bool, firedAlerts []Entry) *action_kit_api.ActionKitError {
	if (state.ExpectedState == alertNotFired && len(firedAlerts) == 0) ||
		(state.ExpectedState == alertFired && len(firedAlerts) > 0) {
		state.StateCheckSuccess = true
	}

	if state.ExpectedState == alertNotFired && len(firedAlerts) > 0 && state.TriggerTime == 0 {
		state.TriggerTime = firedAlerts[0].Content.TriggerTime
	}

	if completed && !state.StateCheckSuccess {
		var title string
		if state.ExpectedState == alertNotFired {
			triggerTime := time.Unix(state.TriggerTime, 0).UTC().Format(time.RFC3339)
			title = fmt.Sprintf("Alert %q should not have been fired but was at %s.", state.Name, triggerTime)
		} else {
			title = fmt.Sprintf("Alert %q should have been fired but was not.", state.Name)
		}
		return extutil.Ptr(action_kit_api.ActionKitError{
			Title:  title,
			Status: extutil.Ptr(action_kit_api.Failed),
		})
	}

	return nil
}

func toMetrics(alertName string, firedAlerts []Entry, now time.Time) []action_kit_api.Metric {
	var triggerTime string
	var tooltip string
	var state string

	if len(firedAlerts) > 0 {
		triggerTime = time.Unix(firedAlerts[0].Content.TriggerTime, 0).UTC().Format(time.RFC3339)
		tooltip = fmt.Sprintf("Splunk Alert %q fired at %s", alertName, triggerTime)
		state = "success"
	}

	return []action_kit_api.Metric{
		{
			Name: extutil.Ptr(fmt.Sprintf("Splunk Alert %s", alertName)),
			Metric: map[string]string{
				metricId:          alertName,
				metricLabel:       alertName,
				metricState:       state,
				metricTooltip:     tooltip,
				metricTriggerTime: triggerTime,
			},
			Timestamp: now,
			Value:     0,
		},
	}
}
