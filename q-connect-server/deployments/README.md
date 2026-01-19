# Deployment Configurations

## Kubernetes
- `k8s/` - Kubernetes manifests for deploying to a K8s cluster

## Structure

```
deployments/
└── k8s/
    ├── postgres.yaml      # PostgreSQL StatefulSet
    ├── valkey.yaml        # Valkey (Redis) StatefulSet
    ├── pv-postgres.yaml   # Persistent Volume for PostgreSQL
    ├── pv-valkey.yaml     # Persistent Volume for Valkey
    ├── deployment.yaml    # Application deployment
    ├── service.yaml       # Application service
    ├── deploy-all.ps1     # Deployment script
    ├── cleanup.ps1        # Cleanup script
    └── README.md          # Deployment guide
```

## Quick Start

```bash
cd deployments/k8s
./deploy-all.ps1
```
