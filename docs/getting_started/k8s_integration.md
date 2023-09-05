# Kubernetes Integration

If you are running multiple frontends in Kubernetes, the b3scale Operator allows for provisioning of new Frontends through CRD. This is independent of whether or not you run b3scale in Kubernetes or not.
## b3scale operator for Kubernetes

```bash
docker pull ghcr.io/b3scale/b3scale-operator:latest
```

## Deployment

The Kubernetes Deployment for the b3scale Operator is described [in the operators' README](https://github.com/b3scale/b3scale-operator/blob/main/README.md).