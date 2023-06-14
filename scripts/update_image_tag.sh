#!/bin/sh
CURRENT_DIR=$(pwd)
cd k8s/env/prod/$APP_NAME && kustomize edit set image $REPO:prod=$REPO:sha-$SHA
cd $CURRENT_DIR
kustomize build k8s/env/prod/$APP_NAME > deploy.yaml