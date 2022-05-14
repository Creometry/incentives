# incentives
Measures node usage and triggers incentives via smartcontract Oracle

# Running:
if this is ran outside a Kubernetes pod , you need to set the KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT env variables
you can this variables by running :
    kubectl describe svc kubernetes

TODO:
    Short Term:
    * collect metrics across all namespaces
    * collect metrics of all pods and nodes
    * reduce privalage in deployments/creometrics-rbac.yaml
    * determine in which namespace should the service account (rbac) exist ?
    * change rbac service account namespace for creometrics


    Long Term:
    * determine what metrics should be collected