#!/usr/bin/env bash
set -euo pipefail

# Ubuntu 24.04 または Rocky 9/10 に Podman と make を入れ、mise.toml のツール
# (Go、podman-compose)を mise でインストールする。何度実行してもよい。
# mise 自体は事前にインストールしておく。
# 使い方: ./setup-linux.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="${SCRIPT_DIR}/docker-compose.yml"

# パッケージマネージャーのコマンド。check_host が設定する。
PKG_UPDATE=()
PKG_INSTALL=()
OS_ID=""

log() { printf '\033[0;32m==>\033[0m %s\n' "$*"; }
warn() { printf '\033[0;33mwarn:\033[0m %s\n' "$*" >&2; }
die() {
  printf '\033[0;31merror:\033[0m %s\n' "$*" >&2
  exit 1
}

check_host() {
  # mise は実行ユーザーのホームにツールを入れる。
  [[ $EUID -ne 0 ]] || die "run as a regular user; the script calls sudo only where it needs to"
  command -v sudo >/dev/null || die "sudo is required"
  command -v mise >/dev/null ||
    die "mise is required; install it first: https://mise.jdx.dev/getting-started.html"

  [[ -r /etc/os-release ]] || die "/etc/os-release not found; cannot identify the distribution"
  # shellcheck source=/dev/null
  . /etc/os-release
  OS_ID="${ID:-unknown}"

  case "$OS_ID" in
    ubuntu | debian)
      PKG_UPDATE=(sudo apt-get update -qq)
      # `sudo env` なら sudoers の setenv なしで環境変数のリセットを回避できる。
      PKG_INSTALL=(sudo env DEBIAN_FRONTEND=noninteractive apt-get install -y)
      [[ "$OS_ID" != "ubuntu" || "${VERSION_ID:-}" == "24.04" ]] ||
        warn "tested on Ubuntu 24.04, found ${VERSION_ID:-unknown}; continuing"
      ;;
    rocky)
      PKG_UPDATE=(sudo dnf makecache -q)
      PKG_INSTALL=(sudo dnf install -y)
      ;;
    *)
      die "unsupported distribution: ${OS_ID} (expected ubuntu, debian or rocky)"
      ;;
  esac
}

# ensure_commands <command>:<package>...
# 入っているかはパッケージ名ではなくコマンドで確認する。
ensure_commands() {
  local spec cmd missing=()
  for spec in "$@"; do
    cmd="${spec%%:*}"
    command -v "$cmd" >/dev/null 2>&1 || missing+=("${spec#*:}")
  done
  if ((${#missing[@]} == 0)); then
    log "already installed: ${*%%:*}"
    return
  fi
  log "Installing packages: ${missing[*]}"
  "${PKG_UPDATE[@]}"
  "${PKG_INSTALL[@]}" "${missing[@]}"
}

install_mise_tools() {
  log "Installing tools from mise.toml"
  mise trust "${SCRIPT_DIR}/mise.toml"
  (cd "$SCRIPT_DIR" && mise install)
}

check_rootless_prerequisites() {
  local mode

  # rootless の Grafana(UID 472)がバインドマウントの親ディレクトリをたどれる必要がある。
  mode="$(stat -c '%A' "$HOME")"
  if [[ "${mode:9:1}" != "x" ]]; then
    warn "${HOME} is ${mode}: rootless podman cannot traverse it as UID 472."
    warn "Grafana will start with no datasources and no dashboards."
    warn "Fix with 'chmod o+x ${HOME}', or add 'user: \"0\"' to the grafana service."
  fi

  # rootless Podman には subuid/subgid の範囲が必要。
  if ! grep -q "^${USER}:" /etc/subuid 2>/dev/null; then
    warn "no /etc/subuid range for ${USER}: rootless podman will refuse to start."
    warn "Fix with: sudo usermod --add-subuids 100000-165535 --add-subgids 100000-165535 ${USER}"
  fi

  # Rocky のバインドマウントには、compose ファイルの SELinux 用 :z ラベルが必要。
  if command -v getenforce >/dev/null 2>&1 && [[ "$(getenforce)" == "Enforcing" ]]; then
    if ! grep -q ':z$' "$COMPOSE_FILE"; then
      warn "SELinux is enforcing but the volumes in docker-compose.yml carry no :z"
      warn "label; Grafana will not be able to read the mounted dashboards."
    fi
  fi

  # Podman には、compose ファイルでレジストリ名付きのイメージ指定が必要。
  if ! grep -qE '^[[:space:]]*image:[[:space:]]*[^[:space:]/]+\.[^[:space:]/]+/' "$COMPOSE_FILE"; then
    warn "the image in docker-compose.yml is not registry-qualified; podman"
    warn "cannot resolve a short name without unqualified-search-registries."
  fi
}

main() {
  check_host
  log "Detected ${OS_ID} ${VERSION_ID:-unknown}"
  ensure_commands make:make podman:podman
  install_mise_tools
  check_rootless_prerequisites

  cat <<EOF

$(log "Setup complete")

Activate mise in your shell (see https://mise.jdx.dev/getting-started.html) so
go and podman-compose are on PATH, then:

  cd ${SCRIPT_DIR}
  # create .env with PROMETHEUS_URL (see README.md)
  make dev

Grafana: http://localhost:3000
EOF
}

main "$@"
