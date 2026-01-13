#! /bin/bash

set -e

# Fetch latest annotated tag
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

# Replace app version in Helm chart
sed -i -E "s/appVersion:.+/appVersion: \"$LATEST_TAG\"/g" helm/Chart.yaml

# Upgrade Helm release
cp ~/.kube/personal-htz.yaml ~/.kube/config
helm upgrade --install shx helm -n personal -f helm/prod-values.yaml
# kubectl rollout restart deployment shx-sphinx -n personal
kubectl delete pods -l app.kubernetes.io/name=sphinx -n personal