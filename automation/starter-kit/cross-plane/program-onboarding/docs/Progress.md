# Crossplane Program Onboarding - Progress

## Phase-I: Install Crossplane + Apply GCP Provider & Secret

**Status:** Testing

**Scope:**
1. Load `config.json`
2. Connect to Kubernetes cluster (`kind-crossplane-dev`)
3. Helm install Crossplane (skips if already installed)
4. Wait for Crossplane pods to be Running in `crossplane-system`
5. `kubectl apply -f 00-provider.yaml` — installs the GCP Compute provider
6. `kubectl apply -f 01-gcp-secret.yaml` — creates the GCP credentials secret

**Files involved:**
- `main.go`
- `config.json`
- `pkg/helmmanager/helmmanager.go`
- `pkg/kubemanager/client.go`, `apply.go`, `wait.go`
- `pkg/crossplane/config-yamls/gcp/00-provider.yaml`
- `pkg/crossplane/config-yamls/gcp/01-gcp-secret.yaml`

**Pre-requisites:**
- `kind-crossplane-dev` cluster is up and reachable
- `01-gcp-secret.yaml` populated with a valid GCP service account key

---

## Phase-II: ProviderConfig + Provider health check

**Status:** Pending

---

## Phase-III: GKE Migration

**Status:** Pending

**Scope:**
- Switch from local `kind-crossplane-dev` cluster to GCP GKE
- No code changes required — config-only switch

**Steps:**
1. Provision GKE cluster and run `gcloud container clusters get-credentials <cluster> --region <region> --project <project>`
2. Update `config.json`:
   - `kubeconfig.context` → GKE context (e.g. `gke_<project>_<region>_<cluster>`)
   - `helm.cross_plane.context` → same GKE context
   - `kubeconfig.kubeconfigPath` → path to kubeconfig if not using default `~/.kube/config`
3. Populate `01-gcp-secret.yaml` with a real GCP service account key
4. Ensure `gke-gcloud-auth-plugin` is installed on the test machine
5. Re-run Phase-I and Phase-II steps against the GKE cluster

**Pre-requisites:**
- GKE cluster is up and reachable
- `gcloud` CLI configured with appropriate project and permissions
- `gke-gcloud-auth-plugin` installed (`gcloud components install gke-gcloud-auth-plugin`)
