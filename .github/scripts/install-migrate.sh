#!/usr/bin/env bash
# Install pinned golang-migrate CLI for CI.
# Prefer `go install` (module proxy). Fall back to a retried, verified
# GitHub release tarball so one flaky assets.githubusercontent.com
# download cannot fail the job.
set -euo pipefail

MIGRATE_VERSION="${MIGRATE_VERSION:-v4.17.0}"
INSTALL_DIR="${MIGRATE_INSTALL_DIR:-/usr/local/bin}"
ARCHIVE="migrate.linux-amd64.tar.gz"
RELEASE_URL="https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/${ARCHIVE}"
CHECKSUM_URL="https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/sha256sum.txt"

die() {
  echo "ERROR: $*" >&2
  exit 1
}

place_binary() {
  local src="$1"
  [[ -f "${src}" ]] || return 1
  if [[ -w "${INSTALL_DIR}" ]]; then
    cp "${src}" "${INSTALL_DIR}/migrate"
  else
    sudo cp "${src}" "${INSTALL_DIR}/migrate"
  fi
  if [[ -w "${INSTALL_DIR}/migrate" ]]; then
    chmod +x "${INSTALL_DIR}/migrate"
  else
    sudo chmod +x "${INSTALL_DIR}/migrate"
  fi
}

install_via_go() {
  command -v go >/dev/null 2>&1 || return 1
  echo "Installing migrate ${MIGRATE_VERSION} via go install (postgres tag)..."
  go install -tags 'postgres' "github.com/golang-migrate/migrate/v4/cmd/migrate@${MIGRATE_VERSION}" || return 1
  local src
  src="$(go env GOPATH)/bin/migrate"
  place_binary "${src}" || return 1
  echo "Installed migrate via go install -> ${INSTALL_DIR}/migrate"
}

download() {
  local url="$1"
  local dest="$2"
  echo "Downloading ${url}"
  curl --retry 5 --retry-all-errors --retry-delay 2 --retry-max-time 120 \
    -fsSL --connect-timeout 15 --max-time 180 \
    -o "${dest}" "${url}"
}

install_via_release() {
  local tmp
  tmp="$(mktemp -d)"
  # shellcheck disable=SC2064
  trap "rm -rf '${tmp}'" RETURN

  download "${RELEASE_URL}" "${tmp}/${ARCHIVE}" || \
    die "failed to download migrate ${MIGRATE_VERSION} from GitHub releases after retries: ${RELEASE_URL}"

  # Reject HTML error pages / truncated payloads before tar.
  if ! gzip -t "${tmp}/${ARCHIVE}" 2>/dev/null; then
    echo "Downloaded file is not a valid gzip archive:" >&2
    file "${tmp}/${ARCHIVE}" >&2 || true
    head -c 300 "${tmp}/${ARCHIVE}" >&2 || true
    echo >&2
    die "migrate archive failed verification (not a valid .tar.gz)"
  fi

  if download "${CHECKSUM_URL}" "${tmp}/sha256sum.txt"; then
    if grep -E "[[:space:]]${ARCHIVE}$" "${tmp}/sha256sum.txt" > "${tmp}/expected.sha"; then
      (cd "${tmp}" && sha256sum -c expected.sha) || die "sha256 mismatch for ${ARCHIVE}"
    else
      echo "Warning: sha256sum.txt did not list ${ARCHIVE}; relying on gzip check"
    fi
  else
    echo "Warning: could not fetch sha256sum.txt; relying on gzip integrity check"
  fi

  tar -xzf "${tmp}/${ARCHIVE}" -C "${tmp}"
  [[ -f "${tmp}/migrate" ]] || die "archive did not contain migrate binary; contents: $(ls -la "${tmp}")"

  place_binary "${tmp}/migrate" || die "failed to install migrate into ${INSTALL_DIR}"
  echo "Installed migrate via release tarball -> ${INSTALL_DIR}/migrate"
}

if ! install_via_go; then
  echo "go install unavailable or failed; falling back to GitHub release tarball"
  install_via_release
fi

command -v migrate >/dev/null 2>&1 || die "migrate is not on PATH after install (${INSTALL_DIR})"
echo -n "migrate version: "
migrate -version
