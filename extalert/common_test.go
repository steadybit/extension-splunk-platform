// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2025 Steadybit GmbH

package extalert

import "context"

type MockSplunkClient struct {
	response []Entry
	err      error
}

func (c MockSplunkClient) Alerts(_ context.Context) ([]Entry, error) {
	return c.response, c.err
}

func (c MockSplunkClient) FiredAlerts(ctx context.Context, alertUrl string) ([]Entry, error) {
	return c.response, c.err
}
