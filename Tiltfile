# Deploy
docker_compose('docker-compose.yml')

# Configure live update for the banco-digital service
docker_build('red-velvet-workspace_banco-digital', '.',
    live_update=[
        sync('.', '/app'),
        run('go mod download', trigger=['go.mod', 'go.sum']),
        run('go build -o banco-digital ./cmd/main.go', trigger=['./cmd/main.go', './internal']),
    ]
)

# Resources