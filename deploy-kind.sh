#!/bin/sh

# this script is used to allow deployment of container images to kind whithout use of online registries

docker build -t creometrics:0.1 .
kind load docker-image creometrics:0.1
kubectl delete -f deployments/creometrics-deployment.yaml
kubectl apply -f deployments/creometrics-deployment.yaml
kubectl get all