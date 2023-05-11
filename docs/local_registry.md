# vHive local registry guide

To avoid bottlenecks, it is possible to use a local registry to store images. This registry is reachable at *docker-registry.registry.svc.cluster.local:5000*.
## Pulling images to the local registry
1. Run skopeo container in a pod

   ` kubectl apply -f configs/registry/skopeo.yaml`
2. Execute copy command in pod skopeo

   Example:
   ```bash
   kubectl exec skopeo -- skopeo copy docker://docker.io/vhiveease/helloworld:var_workload docker://docker-registry.registry.svc.cluster.local:5000/vhiveease/helloworld:var_workload --dest-tls-verify=false
   ```

## Using the local registry

Once the desired images are available at the local registry, it can be used in function deployment by specifying the registry in the image name. When no registry is specified, the registry defaults to *docker.io*.

Example: `docker-registry.registry.svc.cluster.local:5000/vhiveease/helloworld:var_workload`

