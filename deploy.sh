#!/usr/bin/env bash
export APP=orion
export VERSION="$2"

export DOCKERIMAGE=${DOCKERIMAGE:-quay.io/syncano/orion}
TARGET="$1"

usage() { echo "* Usage: $0 <environment> <version> [--skip-push]" >&2; exit 1; }
[[ -n $TARGET ]] || usage
[[ -n $VERSION ]] || usage

set -euo pipefail

if ! command -v > /dev/null; then
    echo "! kubectl not installed" >&2; exit 1
fi

if [[ ! -f "deploy/env/${TARGET}.env" ]]; then
    echo "! environment ${TARGET} does not exist in deploy/env/"; exit 1
fi

# Parse arguments.
PUSH=true
for PARAM in "${@:3}"; do
    case $PARAM in
        --skip-push)
          PUSH=false
          ;;
        *)
          usage
          ;;
    esac
done

envsubst() {
    for var in $(compgen -e); do
        echo "$var: \"${!var//\"/\\\"}\""
    done | PYTHONWARNINGS=ignore jinja2 "$1"
}


echo "* Starting deployment for $TARGET at $VERSION."

# Setup environment variables.
set -a
# shellcheck disable=SC1090
source deploy/env/"${TARGET}".env
set +a
BUILDTIME=$(date +%Y-%m-%dt%H%M)
export BUILDTIME


# Push docker image.
if $PUSH; then
	echo "* Tagging $DOCKERIMAGE $VERSION."
	docker tag "$DOCKERIMAGE" "$DOCKERIMAGE":"$VERSION"

	echo "* Pushing $DOCKERIMAGE:$VERSION."
	docker push "$DOCKERIMAGE":"$VERSION"
fi


# Create configmap.
echo "* Updating ConfigMap."
CONFIGMAP="apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: ${APP}\ndata:\n"
while read -r line; do
    if [[ -n "${line}" && "${line}" != *"#"* ]]; then
        CONFIGMAP+="  ${line%%=*}: \"${line#*=}\"\n"
    fi
done < deploy/env/"${TARGET}".env
echo -e "$CONFIGMAP" | kubectl apply -f -


# Create secrets.
echo "* Updating Secrets."
SECRETS="apiVersion: v1\nkind: Secret\nmetadata:\n  name: ${APP}\ntype: Opaque\ndata:\n"
while read -r line; do
    if [[ -n "${line}" && "${line}" != *"#"* ]]; then
        SECRETS+="  ${line%%=*}: $(echo -n "${line#*=}" | base64 | tr -d '\n')\n"
    fi
done < deploy/env/"${TARGET}".secrets.unenc
echo -e "$SECRETS" | kubectl apply -f -


# Deploy server
REPLICAS=$(kubectl get deployment/orion-server -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "${SERVER_MIN}")
export REPLICAS
echo "* Deploying Server replicas=${REPLICAS}."
envsubst deploy/yaml/server-deployment.yml.j2 | kubectl apply -f -
envsubst deploy/yaml/server-hpa.yml.j2 | kubectl apply -f -

echo "* Deploying Server Service."
envsubst deploy/yaml/server-service.yml.j2 | kubectl apply -f -

kubectl rollout status deployment/orion-server
