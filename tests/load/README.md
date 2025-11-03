# K6 Load Tests - Quick Reference

## Quick Start

```bash
# 1. Install metrics-server (required for HPA)
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# 2. Run integrated test on Kubernetes local
./run-k8s-local.sh integrated

# 3. Monitor scaling in another terminal
./monitor-scaling.sh
```

## Available Tests

- `auth` - Auth service (200 VUs max)
- `product` - Product service (400 VUs max)
- `order` - Order service (1000 VUs max)
- `ui` - UI service (250 VUs max)
- `integrated` - All services (500 VUs max) **← Recommended**
- `all` - Run all tests sequentially

## Scripts

| Script | Description |
|--------|-------------|
| `run-k8s-local.sh [test]` | Run tests on local Kubernetes |
| `monitor-scaling.sh` | Monitor HPA and pods in real-time |

## Configuration

Copy and customize:
```bash
cp .env.local.example .env.local
cp .env.eks.example .env.eks
```

## Full Documentation

See [docs/LOAD_TESTING.md](../../docs/LOAD_TESTING.md) for complete guide.
