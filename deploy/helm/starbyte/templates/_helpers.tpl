{{- /* StarByte Helm helpers — 硬化自 #145 */ -}}

{{/*
标准 fullname：release 名已包含 chart 名时不再重复拼接。
helm install starbyte ...  → starbyte-backend，而不是 starbyte-starbyte-backend
*/}}
{{- define "starbyte.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "starbyte.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "starbyte.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{ include "starbyte.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: starbyte
{{- end }}

{{- define "starbyte.selectorLabels" -}}
app.kubernetes.io/name: {{ include "starbyte.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
镜像引用：repository + tag 必须拆开。
若 tag 含 "/" 或看起来像完整镜像 URL，直接 fail，避免重蹈 #145
`--set backend.image.tag=ghcr.io/.../starbyte-backend:latest` 的覆写事故。
*/}}
{{- define "starbyte.image" -}}
{{- $repo := required "image.repository is required" .repository -}}
{{- $tag := required "image.tag is required" .tag -}}
{{- if or (contains "/" ($tag | toString)) (contains ":" ($tag | toString)) }}
{{- fail (printf "image.tag 只能是 tag（如 1.0.0 / sha-abc123），不能是完整镜像 URL：%s" $tag) }}
{{- end }}
{{- printf "%s:%s" $repo $tag -}}
{{- end }}

{{- define "starbyte.storageClass" -}}
{{- $local := .local -}}
{{- $global := .global -}}
{{- if $local -}}
{{- $local -}}
{{- else if $global -}}
{{- $global -}}
{{- else -}}
{{- "" -}}
{{- end -}}
{{- end }}

{{- define "starbyte.backend.fullname" -}}
{{- printf "%s-backend" (include "starbyte.fullname" .) -}}
{{- end }}

{{- define "starbyte.frontend.fullname" -}}
{{- printf "%s-frontend" (include "starbyte.fullname" .) -}}
{{- end }}

{{- define "starbyte.postgres.fullname" -}}
{{- printf "%s-postgres" (include "starbyte.fullname" .) -}}
{{- end }}

{{- define "starbyte.redis.fullname" -}}
{{- printf "%s-redis" (include "starbyte.fullname" .) -}}
{{- end }}

{{- define "starbyte.minio.fullname" -}}
{{- printf "%s-minio" (include "starbyte.fullname" .) -}}
{{- end }}

{{- define "starbyte.secret.fullname" -}}
{{- printf "%s-secret" (include "starbyte.fullname" .) -}}
{{- end }}

{{- define "starbyte.configmap.fullname" -}}
{{- printf "%s-config" (include "starbyte.fullname" .) -}}
{{- end }}

{{- define "starbyte.ingress.fullname" -}}
{{- printf "%s-ingress" (include "starbyte.fullname" .) -}}
{{- end }}
