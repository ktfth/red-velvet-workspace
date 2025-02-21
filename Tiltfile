# Build
docker_build('banco-digital', '.',
    live_update=[
        sync('.', '/app'),
        run('go mod download', trigger=['go.mod', 'go.sum']),
        run('go build -o banco-digital ./cmd/main.go', trigger=['./cmd/main.go', './internal']),
    ]
)

# Deploy
k8s_yaml(['k8s/deployment.yaml'])

# Port forward
k8s_resource('banco-digital', port_forwards='8080:80') 