echo "========= repo add crossplane-stable ========="
helm repo add crossplane-stable https://charts.crossplane.io/stable

echo "========= repo update ========="
helm repo update

echo "========= install crossplane ========="
helm install crossplane crossplane-stable/crossplane \
  --namespace crossplane-system \
  --create-namespace \
  --kube-context kind-crossplane-dev

echo "========= Sleep for 5 seconds. Allow cross-plane pods to come up ========="
sleep 5

echo "========= Check if the installation is RUNNING ========="
kubectl get pods -n crossplane-system --context kind-crossplane-dev

echo "========= Checking the current context ========="
kubectl config current-context

echo "========= Apply 00-provider.yaml ========="
kubectl apply -f 00-provider.yaml --context kind-crossplane-dev

echo "========= Apply 01-gcp-secret.yaml ========="
kubectl apply -f 01-gcp-secret.yaml --context kind-crossplane-dev

echo "========= Check whether services are operational ========="
kubectl get providers --context kind-crossplane-dev

echo "========= Check of both services are healthy =========" 
#
#NAME                          INSTALLED   HEALTHY   PACKAGE                                               AGE
#provider-gcp-compute          True        True      xpkg.upbound.io/upbound/provider-gcp-compute:v1.8.1   6m15s
#upbound-provider-family-gcp   True        True      xpkg.upbound.io/upbound/provider-family-gcp:v2.5.0    5m39s

echo "========= After applying the last yaml ========="
kubectl get instance.compute.gcp.upbound.io demo-vm   -o wide   --context kind-crossplane-dev
