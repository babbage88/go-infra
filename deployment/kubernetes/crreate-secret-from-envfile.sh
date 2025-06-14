#!/bin/bash
kubectl delete secret goinfra-app-secrets
kubectl create secret generic goinfra-app-secrets --from-env-file=k8s.env
