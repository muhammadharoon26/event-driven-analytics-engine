Write-Host "Starting Docker containers..."
docker-compose up -d redpanda postgres redpanda-init data-processor

Write-Host "Starting Backend (Ingestion API) in a new window..."
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd ingestion-api; go run main.go"

Write-Host "Starting Frontend in a new window..."
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd frontend; npm run dev"

# Write-Host "Starting Data Processor in a new window..."
# Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd data-processor; python main.py"

Write-Host "All services have been started! You can find them in the newly opened terminal windows."
