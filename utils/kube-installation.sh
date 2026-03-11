#!/usr/bin/env bash
set -Eeuo pipefail

export DEBIAN_FRONTEND=noninteractive

# ============================================================
# Non-interactive install:
#   - Docker
#   - kubectl
#   - kind
#   - helm
#   - Google Cloud CLI
#
# Optional env vars:
#   K8S_CHANNEL=v1.35
#   KIND_VERSION=v0.31.0
#   CLUSTER_NAME=demo
#   CREATE_KIND_CLUSTER=true
#   GCP_PROJECT_ID=my-project-id
#   GCP_SA_KEY_FILE=/absolute/path/to/service-account.json
# ============================================================

K8S_CHANNEL="${K8S_CHANNEL:-v1.35}"
KIND_VERSION="${KIND_VERSION:-v0.31.0}"
CLUSTER_NAME="${CLUSTER_NAME:-demo}"
CREATE_KIND_CLUSTER="${CREATE_KIND_CLUSTER:-false}"

INSTALL_DOCKER="${INSTALL_DOCKER:-true}"
INSTALL_KUBECTL="${INSTALL_KUBECTL:-true}"
INSTALL_KIND="${INSTALL_KIND:-true}"
INSTALL_HELM="${INSTALL_HELM:-true}"
INSTALL_GCLOUD="${INSTALL_GCLOUD:-true}"

GCP_PROJECT_ID="${GCP_PROJECT_ID:-}"
GCP_SA_KEY_FILE="${GCP_SA_KEY_FILE:-}"

log() {
  echo
  echo "============================================================"
  echo "$*"
  echo "============================================================"
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1
}

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    echo "Run this script with sudo or as root."
    exit 1
  fi
}

detect_arch() {
  local arch
  arch="$(dpkg --print-architecture)"
  case "${arch}" in
    amd64|arm64) echo "${arch}" ;;
    *)
      echo "Unsupported architecture: ${arch}"
      exit 1
      ;;
  esac
}

install_base_packages() {
  log "Installing base packages"
  apt-get update -y
  apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    apt-transport-https \
    software-properties-common
  install -m 0755 -d /etc/apt/keyrings
}

install_docker() {
  if need_cmd docker; then
    log "Docker already installed, skipping"
    return
  fi

  log "Installing Docker Engine"
  rm -f /etc/apt/keyrings/docker.gpg
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
    | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
  chmod a+r /etc/apt/keyrings/docker.gpg

  . /etc/os-release
  cat >/etc/apt/sources.list.d/docker.list <<EOF
deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu ${VERSION_CODENAME} stable
EOF

  apt-get update -y
  apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

  systemctl enable docker
  systemctl restart docker

  if [[ -n "${SUDO_USER:-}" ]]; then
    usermod -aG docker "${SUDO_USER}" || true
    echo "Added ${SUDO_USER} to docker group."
    echo "A new login shell is usually needed for group membership to apply."
  fi
}

install_kubectl() {
  if need_cmd kubectl; then
    log "kubectl already installed, skipping"
    return
  fi

  log "Installing kubectl from ${K8S_CHANNEL}"
  rm -f /etc/apt/keyrings/kubernetes-apt-keyring.gpg
  curl -fsSL "https://pkgs.k8s.io/core:/stable:/${K8S_CHANNEL}/deb/Release.key" \
    | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
  chmod 0644 /etc/apt/keyrings/kubernetes-apt-keyring.gpg

  cat >/etc/apt/sources.list.d/kubernetes.list <<EOF
deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/${K8S_CHANNEL}/deb/ /
EOF

  apt-get update -y
  apt-get install -y kubectl
}

install_kind() {
  if need_cmd kind; then
    log "kind already installed, skipping"
    return
  fi

  log "Installing kind ${KIND_VERSION}"
  local arch
  arch="$(detect_arch)"

  curl -fsSL -o /usr/local/bin/kind \
    "https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-linux-${arch}"
  chmod +x /usr/local/bin/kind
}

install_helm() {
  if need_cmd helm; then
    log "Helm already installed, skipping"
    return
  fi

  log "Installing Helm"
  rm -f /etc/apt/keyrings/helm.gpg
  curl -fsSL https://packages.buildkite.com/helm-linux/helm-debian/gpgkey \
    | gpg --dearmor -o /etc/apt/keyrings/helm.gpg
  chmod a+r /etc/apt/keyrings/helm.gpg

  cat >/etc/apt/sources.list.d/helm-stable-debian.list <<EOF
deb [signed-by=/etc/apt/keyrings/helm.gpg] https://packages.buildkite.com/helm-linux/helm-debian/any/ any main
EOF

  apt-get update -y
  apt-get install -y helm
}

install_gcloud() {
  if need_cmd gcloud; then
    log "Google Cloud CLI already installed, skipping"
    return
  fi

  log "Installing Google Cloud CLI"
  rm -f /etc/apt/keyrings/google-cloud.gpg
  curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg \
    | gpg --dearmor -o /etc/apt/keyrings/google-cloud.gpg
  chmod a+r /etc/apt/keyrings/google-cloud.gpg

  cat >/etc/apt/sources.list.d/google-cloud-sdk.list <<EOF
deb [signed-by=/etc/apt/keyrings/google-cloud.gpg] https://packages.cloud.google.com/apt cloud-sdk main
EOF

  apt-get update -y
  apt-get install -y google-cloud-cli
}

configure_gcloud_noninteractive() {
  if ! need_cmd gcloud; then
    return
  fi

  if [[ -n "${GCP_SA_KEY_FILE}" ]]; then
    log "Activating gcloud with service account key"
    if [[ ! -f "${GCP_SA_KEY_FILE}" ]]; then
      echo "GCP_SA_KEY_FILE does not exist: ${GCP_SA_KEY_FILE}"
      exit 1
    fi
    gcloud auth activate-service-account --key-file="${GCP_SA_KEY_FILE}" --quiet
  else
    log "Skipping gcloud auth"
    echo "Set GCP_SA_KEY_FILE=/absolute/path/key.json to authenticate non-interactively."
  fi

  if [[ -n "${GCP_PROJECT_ID}" ]]; then
    log "Setting default GCP project"
    gcloud config set project "${GCP_PROJECT_ID}" --quiet
  fi
}

create_kind_cluster() {
  if [[ "${CREATE_KIND_CLUSTER}" != "true" ]]; then
    log "Skipping kind cluster creation"
    return
  fi

  log "Creating kind cluster: ${CLUSTER_NAME}"

  if kind get clusters | grep -qx "${CLUSTER_NAME}"; then
    echo "Kind cluster '${CLUSTER_NAME}' already exists, skipping creation."
    return
  fi

  kind create cluster --name "${CLUSTER_NAME}"
  kubectl cluster-info
  kubectl get nodes -o wide
}

verify_install() {
  log "Verifying installation"
  docker --version || true
  kubectl version --client || true
  kind --version || true
  helm version || true
  gcloud --version || true
}

main() {
  require_root
  install_base_packages

  [[ "${INSTALL_DOCKER}" == "true" ]] && install_docker
  [[ "${INSTALL_KUBECTL}" == "true" ]] && install_kubectl
  [[ "${INSTALL_KIND}" == "true" ]] && install_kind
  [[ "${INSTALL_HELM}" == "true" ]] && install_helm
  [[ "${INSTALL_GCLOUD}" == "true" ]] && install_gcloud

  configure_gcloud_noninteractive
  verify_install
  create_kind_cluster

  log "Done"
  echo "Installed without UI prompts."
  echo "If Docker was newly installed, open a new login shell before using docker as non-root."
}

main "$@"
