#!/bin/bash

echo "Starting Docker containers..."
docker-compose up -d

echo "Starting Backend (Ingestion API) in a new window..."
# Opens a new cmd window, navigates to the backend, and runs it
start cmd /k "cd ingestion-api && go run main.go"

echo "Starting Frontend in a new window..."
# Opens a new cmd window, navigates to the frontend, and runs it
start cmd /k "cd frontend && npm run dev"

# Uncomment below if you also want to start the Python data processor
# echo "Starting Data Processor in a new window..."
# start cmd /k "cd data-processor && python main.py"

echo "All services have been started! You can find them in the newly opened terminal windows."
