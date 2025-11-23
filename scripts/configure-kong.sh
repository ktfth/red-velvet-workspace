#!/bin/sh

# Wait for Kong to be ready
echo "Waiting for Kong to be ready..."
while ! curl -s http://kong:8001 > /dev/null; do
    sleep 1
done

# Helper function to delete all plugins
delete_all_plugins() {
    echo "Deleting all plugins..."
    plugin_ids=$(curl -s http://kong:8001/plugins | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    for id in $plugin_ids; do
        curl -X DELETE http://kong:8001/plugins/$id
    done
}

delete_all_plugins

# Delete existing service if it exists
curl -i -X DELETE http://kong:8001/services/banco-digital

# Create a service for banco-digital
curl -i -X POST http://kong:8001/services \
  --data name=banco-digital \
  --data url=http://red-velvet-workspace-banco-digital-1:8080

# Create a route for the service
curl -i -X POST http://kong:8001/services/banco-digital/routes \
  --data paths[]=/api \
  --data strip_path=true

# Enable CORS plugin globally
curl -i -X POST http://kong:8001/plugins \
  --data name=cors \
  --data config.origins="*" \
  --data config.methods[]=GET \
  --data config.methods[]=POST \
  --data config.methods[]=PUT \
  --data config.methods[]=PATCH \
  --data config.methods[]=DELETE \
  --data config.methods[]=OPTIONS \
  --data config.methods[]=HEAD \
  --data config.headers[]=Accept \
  --data config.headers[]=Accept-Version \
  --data config.headers[]=Content-Length \
  --data config.headers[]=Content-MD5 \
  --data config.headers[]=Content-Type \
  --data config.headers[]=Date \
  --data config.headers[]=X-Auth-Token \
  --data config.headers[]=Authorization \
  --data config.exposed_headers[]=X-Auth-Token \
  --data config.credentials=true \
  --data config.max_age=3600

# JWT IS DISABLED FOR NOW TO FIX FRONTEND ISSUES
# Enable JWT plugin globally
# curl -i -X POST http://kong:8001/plugins \
#   --data name=jwt \
#   --data config.claims_to_verify=exp \
#   --data config.run_on_preflight=false

# Delete existing consumer if it exists
curl -i -X DELETE http://kong:8001/consumers/banco-digital-app

# Create a consumer
curl -i -X POST http://kong:8001/consumers \
  --data username=banco-digital-app

# Create JWT credentials for the consumer
curl -i -X POST http://kong:8001/consumers/banco-digital-app/jwt \
  --data key=37U4VjFAXAItqfppIrZcsfhMBm5hFoxL \
  --data secret=your-256-bit-secret \
  --data algorithm=HS256

echo "Kong configuration completed!"
