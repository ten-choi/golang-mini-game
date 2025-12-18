# Cleanup Kubernetes resources

Write-Host "🧹 Cleaning up Kubernetes resources..." -ForegroundColor Cyan

Write-Host "`n🗑️ Deleting StatefulSets..." -ForegroundColor Yellow
kubectl delete statefulset mongo -n data --ignore-not-found
kubectl delete statefulset valkey -n data --ignore-not-found

Write-Host "`n🗑️ Deleting Services..." -ForegroundColor Yellow
kubectl delete service mongo -n data --ignore-not-found
kubectl delete service valkey -n data --ignore-not-found

Write-Host "`n🗑️ Deleting PVCs..." -ForegroundColor Yellow
kubectl delete pvc --all -n data

Write-Host "`n🗑️ Deleting PVs..." -ForegroundColor Yellow
kubectl delete pv mongo-pv-0 --ignore-not-found
kubectl delete pv valkey-pv-0 --ignore-not-found

Write-Host "`n🗑️ Deleting namespace..." -ForegroundColor Yellow
kubectl delete namespace data --ignore-not-found

Write-Host "`n✅ Cleanup completed!" -ForegroundColor Green
