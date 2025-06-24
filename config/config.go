// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2025 Steadybit GmbH

package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

type Specification struct {
	AccessToken                      string   `json:"accessToken" split_words:"true" required:"true"`
	ApiBaseUrl                       string   `json:"apiBaseUrl" split_words:"true" required:"true"`
	DiscoveryAttributesExcludesAlert []string `json:"discoveryAttributesExcludesAlert" split_words:"true" required:"false"`
	InsecureSkipVerify               bool     `json:"insecureSkipVerify" split_words:"true" default:"false"`
}

var (
	Config Specification
)

func ParseConfiguration() {
	err := envconfig.Process("steadybit_extension", &Config)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to parse configuration from environment.")
	}
}
