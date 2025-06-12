{{/* vim: set filetype=mustache: */}}

{{- define "splunk.secret.name" -}}
{{- default "steadybit-extension-splunk-platform" .Values.splunk.existingSecret -}}
{{- end -}}
