# Crossplane on kind: Create and Delete a GCP VM

This guide documents the end-to-end steps to:

0. Existing scripts (from )
   - helm-cross-plane-install.sh - Install the required packages
   - cleanup-crossplane-gcp.sh   - Clean up th configured resources
1. Create a local `kind` cluster
2. Install Crossplane
3. Apply Crossplane YAMLs to create GCP resources:
   - VPC
   - Subnetwork
   - Firewall
   - VM Instance
4. Verify the resource status
5. Delete the resources cleanly

---

## Prerequisites

Make sure the following are installed on the machine:

- Docker
- kubectl
- kind
- helm
- Google Cloud CLI (`gcloud`)

You also need:

- A valid GCP project
- A GCP service account JSON key with enough permissions to create:
  - VPC network
  - subnet
  - firewall rules
  - VM instance
- The following YAML files in one directory:
  - `00-provider.yaml`
  - `01-gcp-secret.yaml`
  - `02-providerconfig.yaml`
  - `10-network.yaml`
  - `20-subnetwork.yaml`
  - `30-firewall-ssh-http.yaml`
  - `40-instance-vm.yaml`

---

## 1. Create the kind cluster

Create a `kind-config.yaml` file:

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
```

Create the cluster:

```bash
kind create cluster --name crossplane-dev --config kind-config.yaml
```

Verify cluster access:

```bash
kubectl cluster-info --context kind-crossplane-dev
kubectl get nodes --context kind-crossplane-dev
```

---

## 2. Install Crossplane

Add the Crossplane Helm repository:

```bash
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update
```

Install Crossplane into the `crossplane-system` namespace:

```bash
helm install crossplane crossplane-stable/crossplane \
  --namespace crossplane-system \
  --create-namespace \
  --kube-context kind-crossplane-dev
```

Verify Crossplane pods:

```bash
kubectl get pods -n crossplane-system --context kind-crossplane-dev
```

Wait until the Crossplane pods are in `Running` state.

---

## 3. Prepare the GCP credentials secret

Create a base64 version of the service account JSON file:

```bash
base64 -w0 gcp-sa.json
```

Put that output into `01-gcp-secret.yaml` under:

```yaml
data:
  my-gcp-secret: <BASE64_OF_GCP_SERVICE_ACCOUNT_JSON>
```

Also update these values in the YAMLs wherever needed:

- `YOUR_GCP_PROJECT_ID`
- region / zone if you want something other than the defaults

---

## 4. Apply the Crossplane manifests

Apply in this exact order:

```bash
kubectl apply -f 00-provider.yaml --context kind-crossplane-dev
kubectl apply -f 01-gcp-secret.yaml --context kind-crossplane-dev
kubectl apply -f 02-providerconfig.yaml --context kind-crossplane-dev
kubectl apply -f 10-network.yaml --context kind-crossplane-dev
kubectl apply -f 20-subnetwork.yaml --context kind-crossplane-dev
kubectl apply -f 30-firewall-ssh-http.yaml --context kind-crossplane-dev
kubectl apply -f 40-instance-vm.yaml --context kind-crossplane-dev
```

---

## 5. Verify provider installation

Check provider and provider revisions:

```bash
kubectl get providers.pkg.crossplane.io --context kind-crossplane-dev
kubectl get providerrevisions.pkg.crossplane.io --context kind-crossplane-dev
```

Check Crossplane system pods:

```bash
kubectl get pods -n crossplane-system --context kind-crossplane-dev
```

---

## 6. Verify GCP resource creation from Kubernetes

Check all managed resources:

```bash
kubectl get network.compute.gcp.upbound.io --context kind-crossplane-dev
kubectl get subnetwork.compute.gcp.upbound.io --context kind-crossplane-dev
kubectl get firewall.compute.gcp.upbound.io --context kind-crossplane-dev
kubectl get instance.compute.gcp.upbound.io --context kind-crossplane-dev
```

Describe the VM resource:

```bash
kubectl describe instance.compute.gcp.upbound.io demo-vm --context kind-crossplane-dev
```

Watch the VM until it becomes ready:

```bash
kubectl get instance.compute.gcp.upbound.io demo-vm -w --context kind-crossplane-dev
```

You are looking for a successful state such as:

- `SYNCED=True`
- `READY=True`

You can also inspect the full YAML:

```bash
kubectl get instance.compute.gcp.upbound.io demo-vm -o yaml --context kind-crossplane-dev
```

---

## 7. Verify in GCP

List instances in the project:

```bash
gcloud compute instances list --project YOUR_GCP_PROJECT_ID
```

Describe the VM:

```bash
gcloud compute instances describe demo-vm \
  --zone us-central1-a \
  --project YOUR_GCP_PROJECT_ID
```

You can also verify the network and firewall:

```bash
gcloud compute networks list --project YOUR_GCP_PROJECT_ID
gcloud compute firewall-rules list --project YOUR_GCP_PROJECT_ID
```

---

# Deletion / Cleanup

Delete resources in reverse dependency order.

## 1. Delete the VM first

```bash
kubectl delete -f 40-instance-vm.yaml --context kind-crossplane-dev --ignore-not-found=true
```

## 2. Delete the firewall

```bash
kubectl delete -f 30-firewall-ssh-http.yaml --context kind-crossplane-dev --ignore-not-found=true
```

## 3. Delete the subnet

```bash
kubectl delete -f 20-subnetwork.yaml --context kind-crossplane-dev --ignore-not-found=true
```

## 4. Delete the VPC network

```bash
kubectl delete -f 10-network.yaml --context kind-crossplane-dev --ignore-not-found=true
```

## 5. Delete provider config and secret

```bash
kubectl delete -f 02-providerconfig.yaml --context kind-crossplane-dev --ignore-not-found=true
kubectl delete -f 01-gcp-secret.yaml --context kind-crossplane-dev --ignore-not-found=true
```

If the secret delete times out due to a temporary kind/etcd slowdown, retry directly by name:

```bash
kubectl delete secret gcp-secret -n crossplane-system --context kind-crossplane-dev --ignore-not-found=true
```

## 6. Delete the provider

```bash
kubectl delete -f 00-provider.yaml --context kind-crossplane-dev --ignore-not-found=true
```

## 7. Uninstall Crossplane

```bash
helm uninstall crossplane -n crossplane-system --kube-context kind-crossplane-dev || true
```

## 8. Delete the kind cluster

```bash
kind delete cluster --name crossplane-dev
```

---

## Recommended cleanup verification

Before deleting the kind cluster, confirm the GCP resources are gone:

```bash
gcloud compute instances list --project YOUR_GCP_PROJECT_ID
gcloud compute networks list --project YOUR_GCP_PROJECT_ID
gcloud compute firewall-rules list --project YOUR_GCP_PROJECT_ID
```

Also check from Kubernetes side:

```bash
kubectl get instance.compute.gcp.upbound.io --context kind-crossplane-dev
kubectl get firewall.compute.gcp.upbound.io --context kind-crossplane-dev
kubectl get subnetwork.compute.gcp.upbound.io --context kind-crossplane-dev
kubectl get network.compute.gcp.upbound.io --context kind-crossplane-dev
```

---

## Troubleshooting

### Provider shows `HEALTHY=False`

Check:

```bash
kubectl get providerrevisions --context kind-crossplane-dev
kubectl describe providerrevisions --context kind-crossplane-dev
kubectl get pods -n crossplane-system --context kind-crossplane-dev
kubectl get events -n crossplane-system --sort-by=.metadata.creationTimestamp --context kind-crossplane-dev
```

### VM resource exists in Kubernetes but not in GCP

Describe the VM resource and check events:

```bash
kubectl describe instance.compute.gcp.upbound.io demo-vm --context kind-crossplane-dev
```

Common reasons:

- wrong `projectID`
- invalid service account secret
- missing IAM permissions
- subnet/network not ready
- provider not healthy yet

### Secret deletion times out

This may happen if the kind control-plane or etcd is temporarily overloaded. Retry the delete, or delete the whole kind cluster after confirming GCP resources are already removed.

---

## Summary

Creation flow:

1. Create kind cluster
2. Install Crossplane
3. Apply provider and credentials
4. Apply network, subnet, firewall, VM YAMLs
5. Verify in Kubernetes and GCP

Deletion flow:

1. Delete VM
2. Delete firewall
3. Delete subnet
4. Delete network
5. Delete provider config and secret
6. Delete provider
7. Uninstall Crossplane
8. Delete kind cluster

