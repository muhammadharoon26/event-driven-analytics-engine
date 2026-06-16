Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Event-Driven Analytics Engine" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "[1/3] Starting Docker containers (Redpanda, Postgres, Python Worker, Observability Stack)..." -ForegroundColor Yellow
docker-compose up -d redpanda postgres redpanda-init data-processor otel-collector tempo grafana

Write-Host ""
Write-Host "[2/3] Starting Backend (Go Ingestion API) in a new window..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd ingestion-api; Write-Host 'Starting Go Ingestion API...' -ForegroundColor Green; go run main.go"

Write-Host "[3/3] Starting Frontend (React + Vite) in a new window..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd frontend; Write-Host 'Starting React Frontend...' -ForegroundColor Green; npm run dev"

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  All services started!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "  React Dashboard:  http://localhost:5173" -ForegroundColor White
Write-Host "  Go API:           http://localhost:8080" -ForegroundColor White
Write-Host "  Grafana:          http://localhost:3000" -ForegroundColor White
Write-Host "  Health Check:     http://localhost:8080/health" -ForegroundColor White
Write-Host ""
