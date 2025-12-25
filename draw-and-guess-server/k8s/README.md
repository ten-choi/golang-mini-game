# Kubernetes Deployment Guide

이 프로젝트는 PostgreSQL과 Valkey를 Kubernetes에서 실행합니다.

## 📋 Prerequisites

- Kubernetes cluster (Docker Desktop with Kubernetes, Minikube, 등)
- kubectl CLI tool

## 🚀 Quick Start

### 1. PostgreSQL과 Valkey 배포

```powershell
cd k8s
.\deploy-all.ps1
```

이 스크립트는 다음을 수행합니다:
- `data` namespace 생성
- PostgreSQL StatefulSet 배포 (포트: 5432)
- Valkey StatefulSet 배포 (포트: 6379, NodePort: 30379)
- Persistent Volumes 생성

### 2. Pod 상태 확인

```bash
kubectl get pods -n data
```

모든 Pod가 `Running` 상태가 될 때까지 기다립니다.

### 3. 백엔드 서버 실행

```powershell
cd ..
go run src/main.go
```

### 4. 프론트엔드 서버 실행

```powershell
cd ..\draw-and-guess-client
npm run dev
```

## 🗑️ Cleanup

모든 리소스를 삭제하려면:

```powershell
cd k8s
.\cleanup.ps1
```

## 📊 Services

배포 후 다음 엔드포인트를 사용할 수 있습니다:

| 서비스 | 내부 주소 | 외부 주소 |
|-------|----------|----------|
| PostgreSQL | postgres.data.svc.cluster.local:5432 | localhost:5432 |
| Valkey | valkey.data.svc.cluster.local:6379 | localhost:30379 |

## 🔧 Configuration

`.env` 파일에서 다음 설정을 확인하세요:

```env
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=draw_and_guess_db
SERVER_PORT=8080
VALKEY_ADDR=localhost:30379
```

## 📝 Architecture

```
┌─────────────────────────────────────────────────┐
│           Kubernetes Cluster (data ns)          │
│                                                  │
│  ┌──────────────┐         ┌──────────────┐     │
│  │ PostgreSQL   │         │   Valkey     │     │
│  │ StatefulSet  │         │ StatefulSet  │     │
│  │              │         │              │     │
│  │ Port: 5432   │         │ Port: 6379   │     │
│  │              │         │ NodePort:    │     │
│  │              │         │ 30379        │     │
│  └──────────────┘         └──────────────┘     │
│         ▲                        ▲              │
│         │                        │              │
│  ┌──────┴────────┐        ┌──────┴────────┐    │
│  │ PV (hostPath) │        │ PV (hostPath) │    │
│  │  /tmp/data/   │        │  /tmp/data/   │    │
│  │  postgres     │        │   valkey-0    │    │
│  └───────────────┘        └───────────────┘    │
└─────────────────────────────────────────────────┘
                     ▲
                     │
              ┌──────┴──────┐
              │  Backend    │
              │  Server     │
              │ (Port 8080) │
              └─────────────┘
```

## 🐛 Troubleshooting

### Pod가 Pending 상태인 경우

```bash
kubectl describe pod <pod-name> -n data
```

PV가 제대로 바인딩되었는지 확인:

```bash
kubectl get pv
kubectl get pvc -n data
```

### 백엔드가 Valkey/PostgreSQL에 연결할 수 없는 경우

1. Pod가 Running 상태인지 확인
2. 포트 포워딩 확인
3. `.env` 파일의 설정 확인

### 로그 확인

```bash
# PostgreSQL 로그
kubectl logs postgres-0 -n data

# Valkey 로그
kubectl logs valkey-0 -n data
```

## 📦 Storage

- PostgreSQL: 10Gi (hostPath: /tmp/data/postgres-0)
- Valkey: 5Gi (hostPath: /tmp/data/valkey-0)

데이터는 컨테이너가 재시작되어도 유지됩니다.

## 🔐 Security Notes

**중요**: 이 설정은 개발 환경용입니다. 프로덕션 환경에서는:

1. Secret을 사용하여 비밀번호 관리
2. NetworkPolicy로 네트워크 격리
3. RBAC 설정
4. 실제 스토리지 솔루션 사용 (hostPath 대신)
5. Resource limits 적절히 설정

## 📚 Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
- [Persistent Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
