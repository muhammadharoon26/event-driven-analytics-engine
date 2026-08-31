Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Event-Driven Analytics Engine" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Kill any stale Go API processes on port 8080
$existingPid = (netstat -ano | Select-String ":8080.*LISTENING" | ForEach-Object { ($_ -split '\s+')[-1] } | Select-Object -First 1)
if ($existingPid) {
    Write-Host "  Killing stale process on port 8080 (PID $existingPid)..." -ForegroundColor DarkYellow
    Stop-Process -Id $existingPid -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 1
}

Write-Host "[1/4] Starting Docker containers (Redpanda, Postgres, Python Worker, Observability Stack)..." -ForegroundColor Yellow
# NOTE: ingestion-api is excluded — we run Go locally for hot-reload and debugging.
docker compose up -d redpanda postgres redpanda-init data-processor otel-collector tempo grafana

Write-Host ""
Write-Host "[2/4] Waiting for Kafka topics to be created..." -ForegroundColor Yellow
# redpanda-init polls until the broker answers, then creates the topics and
# exits. Blocking here means the Go API never starts against a broker that has
# no 'user-events' topic — that failure mode returns 500 on every publish with
# "Unknown Topic Or Partition" and is confusing to diagnose.
$initExit = (docker wait redpanda-init 2>$null | Select-Object -Last 1)
if ($initExit -eq "0") {
    Write-Host "  Topics ready." -ForegroundColor Green
} else {
    Write-Host "  WARNING: topic creation failed (exit $initExit)." -ForegroundColor Red
    Write-Host "  Check: docker logs redpanda-init" -ForegroundColor Red
    Write-Host "  Fix:   docker exec redpanda rpk topic create user-events --brokers redpanda:29092" -ForegroundColor Red
}

Write-Host ""
Write-Host "[3/4] Starting Backend (Go Ingestion API) in a new window..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd 'e:\Git-Hub\event-driven-analytics-engine\ingestion-api'; Write-Host '=== Go Ingestion API ===' -ForegroundColor Green; Write-Host 'Building... (first run takes ~30s)' -ForegroundColor DarkYellow; go run main.go"

Write-Host "[4/4] Starting Frontend (React + Vite) in a new window..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd 'e:\Git-Hub\event-driven-analytics-engine\frontend'; Write-Host '=== React Frontend ===' -ForegroundColor Green; npm run dev"

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
Write-Host "  NOTE: Go API takes ~30s to compile on first run." -ForegroundColor DarkGray
Write-Host "        Check its terminal window for progress." -ForegroundColor DarkGray
Write-Host ""
