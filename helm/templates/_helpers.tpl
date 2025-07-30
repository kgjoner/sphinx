{{- define "version.major" -}}
{{- regexFind "^v\\d+" .Chart.AppVersion -}}
{{- end -}}

{{- define "prefix" -}}
{{ .Chart.Name }}-{{ template "version.major" . }}
{{- end -}}

{{- define "api.serviceName" -}}
{{ template "prefix" . }}
{{- end -}}

{{- define "api.configMapName" -}}
{{ template "prefix" . }}-config
{{- end -}}

{{- define "api.secretName" -}}
{{ template "prefix" . }}-secret
{{- end -}}

{{- define "db.serviceName" -}}
{{ template "prefix" . }}-pg
{{- end -}}

{{- define "db.secretName" -}}
{{ template "prefix" . }}-pg-secret
{{- end -}}

{{- define "redis.serviceName" -}}
{{ template "prefix" . }}-redis
{{- end -}}

{{- define "redis.configName" -}}
{{ template "prefix" . }}-redis-config
{{- end -}}



