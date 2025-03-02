import jwt
import time

# Credenciais do Kong
key = "vrIm0jvwvREuZlJVLmhLFso9S8SXJV1Z"
secret = "your-256-bit-secret"

# Criar payload
payload = {
    'iss': key,
    'exp': int(time.time()) + 3600  # Token válido por 1 hora
}

# Gerar token
token = jwt.encode(payload, secret, algorithm='HS256')
print(f"Token JWT: {token}")
