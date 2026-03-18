#! /bin/bash

set -e

ENV=
for arg in "$@"; do
  case $arg in
    --env=*)
      ENV="${arg#*=}"
      shift
      ;;
    --env)
      ENV="$2"
      shift 2
      ;;
    *)
      echo "Unknown argument: $arg"
      ;;
  esac
  shift
done

if [ -z "$ENV" ]; then
  echo "Error: --env argument is required (e.g. --env=staging)"
  exit 1
fi

# TODO: ensure kube context is the right one before applying helm charts

# Upgrade Helm release
helm upgrade --install iam helm -n ${ENV} -f helm/${ENV}-values.yaml