{{- /* StarByte Helm Chart - Issue #102 */ -}}
{{/* 模板辅助函数 */}}

{{/*
生成资源全名,截断至 63 字符以符合 DNS-1123 子域规范
*/}}
{{- define "starbyte.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- $fullname := printf "%s-%s" .Release.Name $name -}}
{{- trunc 63 (trimSuffix "-" $fullname) -}}
{{- end -}}

{{/*
生成带 chart 名称前缀的全名 (用于多 chart 共存场景)
*/}}
{{- define "starbyte.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
统一标签 (labels)
*/}}
{{- define "starbyte.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{ include "starbyte.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: starbyte
{{- end -}}

{{/*
选择标签 (selectorLabels),用于 Deployment/StatefulSet 的 selector
*/}}
{{- define "starbyte.selectorLabels" -}}
app.kubernetes.io/name: {{ include "starbyte.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
storageClass 处理:优先使用组件级 storageClass,否则回退到 global.storageClass
返回空字符串 "" 表示使用集群默认 StorageClass
*/}}
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
{{- end -}}

{{/*
后端服务的全名
*/}}
{{- define "starbyte.backend.fullname" -}}
{{- printf "%s-backend" (include "starbyte.fullname" .) -}}
{{- end -}}

{{/*
前端服务的全名
*/}}
{{- define "starbyte.frontend.fullname" -}}
{{- printf "%s-frontend" (include "starbyte.fullname" .) -}}
{{- end -}}

{{/*
PostgreSQL 的全名
*/}}
{{- define "starbyte.postgres.fullname" -}}
{{- printf "%s-postgres" (include "starbyte.fullname" .) -}}
{{- end -}}

{{/*
Redis 的全名
*/}}
{{- define "starbyte.redis.fullname" -}}
{{- printf "%s-redis" (include "starbyte.fullname" .) -}}
{{- end -}}

{{/*
MinIO 的全名
*/}}
{{- define "starbyte.minio.fullname" -}}
{{- printf "%s-minio" (include "starbyte.fullname" .) -}}
{{- end -}}

{{/*
Secret 的全名
*/}}
{{- define "starbyte.secret.fullname" -}}
{{- printf "%s-secret" (include "starbyte.fullname" .) -}}
{{- end -}}

{{/*
ConfigMap 的全名
*/}}
{{- define "starbyte.configmap.fullname" -}}
{{- printf "%s-config" (include "starbyte.fullname" .) -}}
{{- end -}}

{{/*
Ingress 的全名
*/}}
{{- define "starbyte.ingress.fullname" -}}
{{- printf "%s-ingress" (include "starbyte.fullname" .) -}}
{{- end -}}
