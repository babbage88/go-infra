#!/bin/bash
kubectl create secret generic goinfra-app-secrets --from-env-file=k8s.env -n development
