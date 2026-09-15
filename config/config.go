// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2025 Steadybit GmbH

package config

import (
	"strings"

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

func ValidateConfiguration() {
	// envconfig's `required:"true"` only checks that the variable is *set*: an empty
	// value satisfies it, so the extension would start with a blank configuration and
	// fail much later against the target system. Reject blank values here instead.
	if strings.TrimSpace(Config.AccessToken) == "" {
		log.Fatal().Msg("STEADYBIT_EXTENSION_ACCESS_TOKEN must not be empty.")
	}
	if strings.TrimSpace(Config.ApiBaseUrl) == "" {
		log.Fatal().Msg("STEADYBIT_EXTENSION_API_BASE_URL must not be empty.")
	}
}
