# Changelog

## v1.0.13

- Add a "Fail early" option to the alert status check. When enabled (the default, matching the previous behavior), the "All the time" mode fails as soon as a deviating state is observed. When disabled, the check keeps collecting events for the whole duration and only fails at the end of the step. Only affects the "All the time" mode.
- Fix link in README.md (#43)
- chore(deps): bump github.com/steadybit/action-kit/go/action_kit_sdk
- chore(deps): bump github.com/steadybit/discovery-kit/go/discovery_kit_sdk
- chore(deps): bump github.com/steadybit/extension-kit
- chore(deps): bump go to 1.26.5 (#41)
- chore: add Claude Code workflows (#36)
- chore: silence SonarQube finding on secrets: inherit in Claude workflows
- ci: skip build on .trivyignore.yml-only changes [skip ci]
- feat(alert check): add fail early option (#44)
- fix: guard alert check attributes and fix unbounded alert paging
- fix: guard the alert check against targets missing the name/url attributes instead of panicking
- fix: terminate alert paging on the returned page size instead of the server-reported total, preventing an infinite request loop (and dropped results) when Splunk reports an inaccurate total
- refactor: register extension index via exthttp.RegisterRevisionedHandler (#42)

## v1.0.12


## v1.0.11

- chore(deps): bump alpine from 3.23 to 3.24
- chore(deps): bump golang.org/x/net to v0.55.0 (CVE-2026-39821) (#31)

## v1.0.10

- chore: update to go 1.26.4
- feat: add weekly auto patch-release workflow

## v1.0.9

- Support discovery group attribute via `STEADYBIT_EXTENSION_DISCOVERY_GROUP` env var (or `discovery.group` Helm value) — when set, the extension adds `steadybit.group=<value>` to every discovered target
- Update dependencies

## v1.0.8

- Bump Go to 1.26.3
- Update dependencies

## v1.0.7

- Bump Go to 1.25.9
- Support if-none-match for the extension list endpoint
- Update dependencies

## v1.0.6

- feat(chart): split image.name into image.registry + image.name
- Support global.priorityClassName
- Update alpine packages in Docker image to address CVEs
- Update dependencies

## v1.0.5

- Update dependencies

## v1.0.4

- Update dependencies

## v1.0.3

- Update dependencies

## v1.0.2

 - Update dependencies

## v1.0.1

 - Add support for self-signed certificates
 - Update dependencies

## v1.0.0

 - Initial release
