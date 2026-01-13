{{- define "sphinx.version" -}}
{{- .Values.api.image.tag | default .Chart.AppVersion -}}
{{- end -}}

{{- define "sphinx.version.major" -}}
{{- regexFind "^v\\d+" (include "sphinx.version" .) -}}
{{- end -}}

{{- define "sphinx.prefix" -}}
{{- if contains .Chart.Name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "sphinx.serviceName" -}}
{{ template "sphinx.prefix" . }}
{{- end -}}

{{- define "sphinx.configMapName" -}}
{{ template "sphinx.prefix" . }}-config
{{- end -}}

{{- define "sphinx.secretName" -}}
{{ template "sphinx.prefix" . }}-secret
{{- end -}}

{{- define "sphinx.db.serviceName" -}}
{{ template "sphinx.prefix" . }}-pg
{{- end -}}

{{- define "sphinx.db.secretName" -}}
{{ template "sphinx.prefix" . }}-pg-secret
{{- end -}}

{{- define "sphinx.redis.serviceName" -}}
{{ template "sphinx.prefix" . }}-redis
{{- end -}}

{{- define "sphinx.redis.configName" -}}
{{ template "sphinx.prefix" . }}-redis-config
{{- end -}}
