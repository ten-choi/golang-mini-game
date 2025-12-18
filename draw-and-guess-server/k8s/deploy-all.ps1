# Deploy MongoDB and Valkey to Kubernetes

Write-Host "🚀 Deploying MongoDB and Valkey to Kubernetes..." -ForegroundColor Cyan

# Create data namespace
Write-Host "`n📁 Creating namespace..." -ForegroundColor Yellow
kubectl apply -f mongo.yaml

# Create Persistent Volumes
Write-Host "`n💾 Creating Persistent Volumes..." -ForegroundColor Yellow
kubectl apply -f pv-mongo.yaml
kubectl apply -f pv-valkey.yaml

# Deploy Valkey
Write-Host "`n🔴 Deploying Valkey..." -ForegroundColor Yellow
kubectl apply -f valkey.yaml

# Wait for pods to be ready
Write-Host "`n⏳ Waiting for pods to be ready..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Show status
Write-Host "`n✅ Deployment Status:" -ForegroundColor Green
kubectl get pods -n data
kubectl get svc -n data
kubectl get pv

Write-Host "`n📊 Services:" -ForegroundColor Green
Write-Host "MongoDB: mongo.data.svc.cluster.local:27017"
Write-Host "Valkey: valkey.data.svc.cluster.local:6379"
Write-Host "Valkey NodePort: localhost:30379"

Write-Host "`n✨ Deployment completed!" -ForegroundColor Green
