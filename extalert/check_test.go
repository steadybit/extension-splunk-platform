// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2025 Steadybit GmbH

package extalert

import (
	"errors"
	"github.com/google/uuid"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestAlertCheckAction_Describe_NoError(t *testing.T) {
	action := NewAlertCheckAction(nil)

	description := action.Describe()

	require.NotNil(t, description)
}

func TestAlertCheckAction_Prepare(t *testing.T) {
	action := NewAlertCheckAction(nil)
	state := action.NewEmptyState()
	ctx := t.Context()

	_, err := action.Prepare(ctx, &state, action_kit_api.PrepareActionRequestBody{
		Target: &action_kit_api.Target{
			Attributes: map[string][]string{
				attributeID:   {"id"},
				attributeName: {"name"},
				attributeUrl:  {"url"},
			},
		},
		Config: map[string]interface{}{
			"duration":           1000,
			"expectedState":      alertFired,
			"stateCheckMode":     stateCheckModeAtLeastOnce,
			"checkNewAlertsOnly": true,
		},
		ExecutionId: uuid.UUID{},
	})

	require.NoError(t, err)
	require.Equal(t, "id", state.Id)
	require.Equal(t, "id", state.Id)
	require.Equal(t, "name", state.Name)
	require.Equal(t, "url", state.Url)
	require.NotZero(t, state.Start)
	require.Greater(t, state.End, state.Start)
	require.Equal(t, alertFired, state.ExpectedState)
	require.Equal(t, stateCheckModeAtLeastOnce, state.StateCheckMode)
	require.Equal(t, true, state.CheckNewAlertsOnly)
	require.Equal(t, false, state.StateCheckSuccess)
}

func TestAlertCheckAction_Start_NoError(t *testing.T) {
	action := NewAlertCheckAction(nil)
	state := action.NewEmptyState()
	ctx := t.Context()

	_, err := action.Start(ctx, &state)

	require.NoError(t, err)
}

func TestAlertCheckAction_checkFiredAlerts_atLeastOnce_expectFired_noAlertPresent_running(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{},
	}
	state := AlertCheckState{
		Id:             "alertId",
		Name:           "Alert Name",
		Url:            "http://example.com/alertId",
		Start:          time.Now().Add(-1 * time.Minute),
		End:            time.Now().Add(1 * time.Minute),
		ExpectedState:  alertFired,
		StateCheckMode: stateCheckModeAtLeastOnce,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.False(t, result.Completed)
	require.Nil(t, result.Error)
}

func TestAlertCheckAction_checkFiredAlerts_atLeastOnce_expectFired_noAlertPresent_completed(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{},
	}
	state := AlertCheckState{
		Id:             "alertId",
		Name:           "Alert Name",
		Url:            "http://example.com/alertId",
		Start:          time.Now().Add(-2 * time.Minute),
		End:            time.Now().Add(-1 * time.Minute),
		ExpectedState:  alertFired,
		StateCheckMode: stateCheckModeAtLeastOnce,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.True(t, result.Completed)
	require.NotNil(t, result.Error)
	require.Contains(t, result.Error.Title, `Alert "Alert Name" should have been fired but was not.`)
}

func TestAlertCheckAction_checkFiredAlerts_atLeastOnce_expectFired_withAlert_running(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{
			{
				Content: Content{
					TriggerTime: 946684800,
				},
			},
		},
	}
	now := time.Now()
	state := AlertCheckState{
		Start:          now.Add(-1 * time.Minute),
		End:            now.Add(+1 * time.Minute),
		ExpectedState:  alertFired,
		StateCheckMode: stateCheckModeAtLeastOnce,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.False(t, result.Completed)
	require.Nil(t, result.Error)
}

func TestAlertCheckAction_checkFiredAlerts_atLeastOnce_expectFired_withAlert_completed(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{
			{
				Content: Content{
					TriggerTime: 946684800,
				},
			},
		},
	}
	now := time.Now()
	state := AlertCheckState{
		Start:          now.Add(-2 * time.Minute),
		End:            now.Add(-1 * time.Minute),
		ExpectedState:  alertFired,
		StateCheckMode: stateCheckModeAtLeastOnce,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.True(t, result.Completed)
	require.Nil(t, result.Error)
}

func TestAlertCheckAction_checkFiredAlerts_atLeastOnce_expectNotFired_noAlertPresent_running(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{},
	}
	state := AlertCheckState{
		Id:             "alertId",
		Name:           "Alert Name",
		Url:            "http://example.com/alertId",
		Start:          time.Now().Add(-1 * time.Minute),
		End:            time.Now().Add(1 * time.Minute),
		ExpectedState:  alertNotFired,
		StateCheckMode: stateCheckModeAtLeastOnce,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.False(t, result.Completed)
	require.Nil(t, result.Error)
}

func TestAlertCheckAction_checkFiredAlerts_atLeastOnce_expectNotFired_noAlertPresent_completed(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{},
	}
	state := AlertCheckState{
		Id:             "alertId",
		Name:           "Alert Name",
		Url:            "http://example.com/alertId",
		Start:          time.Now().Add(-2 * time.Minute),
		End:            time.Now().Add(-1 * time.Minute),
		ExpectedState:  alertNotFired,
		StateCheckMode: stateCheckModeAtLeastOnce,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.True(t, result.Completed)
	require.Nil(t, result.Error)
}

func TestAlertCheckAction_checkFiredAlerts_atLeastOnce_expectNotFired_withAlert_running(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{
			{
				Content: Content{
					TriggerTime: 946684800,
				},
			},
		},
	}
	now := time.Now()
	state := AlertCheckState{
		Start:          now.Add(-1 * time.Minute),
		End:            now.Add(+1 * time.Minute),
		ExpectedState:  alertNotFired,
		StateCheckMode: stateCheckModeAtLeastOnce,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.False(t, result.Completed)
	require.Nil(t, result.Error)
}

func TestAlertCheckAction_checkFiredAlerts_atLeastOnce_expectNotFired_withAlert_completed(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{
			{
				Content: Content{
					TriggerTime: 946684800,
				},
			},
		},
	}
	now := time.Now()
	state := AlertCheckState{
		Id:             "alertId",
		Name:           "Alert Name",
		Start:          now.Add(-2 * time.Minute),
		End:            now.Add(-1 * time.Minute),
		ExpectedState:  alertNotFired,
		StateCheckMode: stateCheckModeAtLeastOnce,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.True(t, result.Completed)
	require.NotNil(t, result.Error)
	require.True(t, strings.HasPrefix(result.Error.Title, `Alert "Alert Name" should not have been fired but was at`))
}

func TestAlertCheckAction_checkFiredAlerts_allTheTime_expectFired_noAlertPresent_running(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{},
	}
	state := AlertCheckState{
		Id:             "alertId",
		Name:           "Alert Name",
		Url:            "http://example.com/alertId",
		Start:          time.Now().Add(-1 * time.Minute),
		End:            time.Now().Add(1 * time.Minute),
		ExpectedState:  alertFired,
		StateCheckMode: stateCheckModeAllTheTime,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.False(t, result.Completed)
	require.NotNil(t, result.Error)
	require.Contains(t, result.Error.Title, `Alert "Alert Name" should have been fired all the time but was not.`)
}

func TestAlertCheckAction_checkFiredAlerts_allTheTime_expectFired_noAlertPresent_completed(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{},
	}
	state := AlertCheckState{
		Id:             "alertId",
		Name:           "Alert Name",
		Url:            "http://example.com/alertId",
		Start:          time.Now().Add(-2 * time.Minute),
		End:            time.Now().Add(-1 * time.Minute),
		ExpectedState:  alertFired,
		StateCheckMode: stateCheckModeAllTheTime,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.True(t, result.Completed)
	require.NotNil(t, result.Error)
	require.Contains(t, result.Error.Title, `Alert "Alert Name" should have been fired all the time but was not.`)
}

func TestAlertCheckAction_checkFiredAlerts_allTheTime_expectFired_withAlert_running(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{
			{
				Content: Content{
					TriggerTime: 946684800,
				},
			},
		},
	}
	now := time.Now()
	state := AlertCheckState{
		Start:          now.Add(-1 * time.Minute),
		End:            now.Add(+1 * time.Minute),
		ExpectedState:  alertFired,
		StateCheckMode: stateCheckModeAllTheTime,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.False(t, result.Completed)
	require.Nil(t, result.Error)
}

func TestAlertCheckAction_checkFiredAlerts_allTheTime_expectFired_withAlert_completed(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{
			{
				Content: Content{
					TriggerTime: 946684800,
				},
			},
		},
	}
	now := time.Now()
	state := AlertCheckState{
		Start:          now.Add(-2 * time.Minute),
		End:            now.Add(-1 * time.Minute),
		ExpectedState:  alertFired,
		StateCheckMode: stateCheckModeAllTheTime,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.True(t, result.Completed)
	require.Nil(t, result.Error)
}

func TestAlertCheckAction_checkFiredAlerts_allTheTime_expectNotFired_noAlertPresent_running(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{},
	}
	state := AlertCheckState{
		Id:             "alertId",
		Name:           "Alert Name",
		Url:            "http://example.com/alertId",
		Start:          time.Now().Add(-1 * time.Minute),
		End:            time.Now().Add(1 * time.Minute),
		ExpectedState:  alertNotFired,
		StateCheckMode: stateCheckModeAllTheTime,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.False(t, result.Completed)
	require.Nil(t, result.Error)
}

func TestAlertCheckAction_checkFiredAlerts_allTheTime_expectNotFired_noAlertPresent_completed(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{},
	}
	state := AlertCheckState{
		Id:             "alertId",
		Name:           "Alert Name",
		Url:            "http://example.com/alertId",
		Start:          time.Now().Add(-2 * time.Minute),
		End:            time.Now().Add(-1 * time.Minute),
		ExpectedState:  alertNotFired,
		StateCheckMode: stateCheckModeAllTheTime,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.True(t, result.Completed)
	require.Nil(t, result.Error)
}

func TestAlertCheckAction_checkFiredAlerts_allTheTime_expectNotFired_withAlert_running(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{
			{
				Content: Content{
					TriggerTime: 946684800,
				},
			},
		},
	}
	now := time.Now()
	state := AlertCheckState{
		Id:             "alertId",
		Name:           "Alert Name",
		Start:          now.Add(-1 * time.Minute),
		End:            now.Add(+1 * time.Minute),
		ExpectedState:  alertNotFired,
		StateCheckMode: stateCheckModeAllTheTime,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.False(t, result.Completed)
	require.NotNil(t, result.Error)
	require.True(t, strings.HasPrefix(result.Error.Title, `Alert "Alert Name" should not have been fired but was at`))
}

func TestAlertCheckAction_checkFiredAlerts_allTheTime_expectNotFired_withAlert_completed(t *testing.T) {
	mockClient := MockSplunkClient{
		response: []Entry{
			{
				Content: Content{
					TriggerTime: 946684800,
				},
			},
		},
	}
	now := time.Now()
	state := AlertCheckState{
		Id:             "alertId",
		Name:           "Alert Name",
		Start:          now.Add(-2 * time.Minute),
		End:            now.Add(-1 * time.Minute),
		ExpectedState:  alertNotFired,
		StateCheckMode: stateCheckModeAllTheTime,
	}

	result, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NoError(t, err)
	require.True(t, result.Completed)
	require.NotNil(t, result.Error)
	require.True(t, strings.HasPrefix(result.Error.Title, `Alert "Alert Name" should not have been fired but was at`))
}

func TestAlertCheckAction_checkFiredAlerts_onlyNewAlerts_withOldOne(t *testing.T) {
	now := time.Now()
	mockClient := MockSplunkClient{
		response: []Entry{
			{
				Content: Content{
					TriggerTime: now.Unix(),
				},
			},
		},
	}
	state := AlertCheckState{
		Id:                 "alertId",
		Name:               "Alert Name",
		Start:              now,
		End:                now.Add(1 * time.Minute),
		ExpectedState:      alertNotFired,
		StateCheckMode:     stateCheckModeAllTheTime,
		CheckNewAlertsOnly: true,
	}

	result, _ := checkFiredAlerts(t.Context(), &state, mockClient)

	require.Nil(t, result.Error)
}

func TestAlertCheckAction_checkFiredAlerts_onlyNewAlerts_withNewOne(t *testing.T) {
	now := time.Now()
	mockClient := MockSplunkClient{
		response: []Entry{
			{
				Content: Content{
					TriggerTime: now.Add(1 * time.Second).Unix(),
				},
			},
		},
	}
	state := AlertCheckState{
		Id:                 "alertId",
		Name:               "Alert Name",
		Start:              now,
		End:                now.Add(1 * time.Minute),
		ExpectedState:      alertNotFired,
		StateCheckMode:     stateCheckModeAllTheTime,
		CheckNewAlertsOnly: true,
	}

	result, _ := checkFiredAlerts(t.Context(), &state, mockClient)

	require.NotNil(t, result.Error)
}

func TestAlertCheckAction_checkFiredAlerts_error_on_lookup(t *testing.T) {
	mockClient := MockSplunkClient{
		err: errors.New("lookup error"),
	}
	now := time.Now()
	state := AlertCheckState{
		Id:             "alertId",
		Name:           "Alert Name",
		Start:          now.Add(-2 * time.Minute),
		End:            now.Add(-1 * time.Minute),
		ExpectedState:  alertNotFired,
		StateCheckMode: stateCheckModeAllTheTime,
	}

	_, err := checkFiredAlerts(t.Context(), &state, mockClient)

	require.Error(t, err)
}
