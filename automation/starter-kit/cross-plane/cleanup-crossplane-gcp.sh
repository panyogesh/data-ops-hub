#!/usr/bin/env bash
set -Eeuo pipefail

# ============================================================
# Cleanup Crossplane-managed GCP demo resources
# Optional:
#   DELETE_PROVIDER=true
#   DELETE_CROSSPLANE=true
#   DELETE_KIND_CLUSTER=true
#   KIND_CLUSTER_NAME=crossplane-dev
#   KUBE_CONTEXT=kind-crossplane-dev
# ============================================================

KUBE_CONTEXT="${KUBE_CONTEXT:-kind-crossplane-dev}"
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-crossplane-dev}"

DELETE_PROVIDER="${DELETE_PROVIDER:-false}"
DELETE_CROSSPLANE="${DELETE_CROSSPLANE:-false}"
DELETE_KIND_CLUSTER="${DELETE_KIND_CLUSTER:-false}"

log() {
  echo
  echo "============================================================"
  echo "$*"
  echo "============================================================"
}

k() {
  kubectl --context "${KUBE_CONTEXT}" "$@"
}

delete_if_exists() {
  local file="$1"
  if [[ -f "${file}" ]]; then
    echo "Deleting ${file}"
    k delete -f "${file}" --ignore-not-found=true
  else
    echo "Skipping ${file} (file not found)"
  fi
}

wait_delete_kind() {
  local kind_name="$1"
  if kind get clusters | grep -qx "${kind_name}"; then
    kind delete cluster --name "${kind_name}"
  else
    echo "Kind cluster '${kind_name}' not found, skipping"
  fi
}

show_status() {
  echo
  echo "Remaining Crossplane resources:"
  k get instance.compute.gcp.upbound.io 2>/dev/null || true
  k get firewall.compute.gcp.upbound.io 2>/dev/null || true
  k get subnetwork.compute.gcp.upbound.io 2>/dev/null || true
  k get network.compute.gcp.upbound.io 2>/dev/null || true
  k get providers.pkg.crossplane.io 2>/dev/null || true
}

main() {
  log "Deleting managed GCP resources in dependency order"
  delete_if_exists 40-instance-vm.yaml
  delete_if_exists 30-firewall-ssh-http.yaml
  delete_if_exists 20-subnetwork.yaml
  delete_if_exists 10-network.yaml

  log "Waiting briefly for cloud resource deletions to propagate"
  sleep 10

  if [[ "${DELETE_PROVIDER}" == "true" ]]; then
    log "Deleting ProviderConfig, secret, and provider"
    delete_if_exists 02-providerconfig.yaml
    delete_if_exists 01-gcp-secret.yaml
    delete_if_exists 00-provider.yaml
  fi

  if [[ "${DELETE_CROSSPLANE}" == "true" ]]; then
    log "Uninstalling Crossplane Helm release"
    helm uninstall crossplane -n crossplane-system --kube-context "${KUBE_CONTEXT}" || true
  fi

  show_status

  if [[ "${DELETE_KIND_CLUSTER}" == "true" ]]; then
    log "Deleting kind cluster"
    wait_delete_kind "${KIND_CLUSTER_NAME}"
  fi

  log "Cleanup complete"
}
main "$@"
