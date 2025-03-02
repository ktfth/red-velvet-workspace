#!/bin/sh

# Wait for Kong to be ready
echo "Waiting for Kong to be ready..."
while ! curl -s http://kong:8001 > /dev/null; do
    sleep 1
done

# Delete existing service if it exists
curl -i -X DELETE http://kong:8001/services/banco-digital

# Create a service for banco-digital
curl -i -X POST http://kong:8001/services \
  --data name=banco-digital \
  --data url=http://red-velvet-workspace-banco-digital-1:8080

# Delete existing route if it exists
curl -i -X DELETE http://kong:8001/services/banco-digital/routes

# Create a route for the service
curl -i -X POST http://kong:8001/services/banco-digital/routes \
  --data paths[]=/api \
  --data strip_path=true

# Delete existing JWT plugin if it exists
curl -i -X DELETE http://kong:8001/plugins \
  --data name=jwt

# Enable JWT plugin
curl -i -X POST http://kong:8001/plugins \
  --data name=jwt \
  --data config.claims_to_verify=exp

# Delete existing consumer if it exists
curl -i -X DELETE http://kong:8001/consumers/banco-digital-app

# Create a consumer
curl -i -X POST http://kong:8001/consumers \
  --data username=banco-digital-app

# Create JWT credentials for the consumer
curl -i -X POST http://kong:8001/consumers/banco-digital-app/jwt \
  --data algorithm=HS256 \
  --data secret=your-256-bit-secret

echo "Kong configuration completed!"
