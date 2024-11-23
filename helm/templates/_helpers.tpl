{{- define "api.serviceName" -}}
{{ .Chart.Name }}
{{- end -}}

{{- define "api.configMapName" -}}
{{ .Chart.Name }}-configmap
{{- end -}}

{{- define "api.secretName" -}}
{{ .Chart.Name }}-secret
{{- end -}}

{{- define "db.serviceName" -}}
{{ .Chart.Name }}-db
{{- end -}}

{{- define "db.secretName" -}}
{{ .Chart.Name }}-db-secret
{{- end -}}


