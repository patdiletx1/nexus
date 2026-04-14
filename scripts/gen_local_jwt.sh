#!/usr/bin/env bash

set -euo pipefail

JWT_SECRET="${JWT_SECRET:-local-dev-secret}"
JWT_SUB="${JWT_SUB:-user-local-1}"
JWT_ROLE="${JWT_ROLE:-owner}"
JWT_COMPANY_ID="${JWT_COMPANY_ID:-company-local-1}"
JWT_EXP_SECONDS="${JWT_EXP_SECONDS:-3600}"

python3 - "${JWT_SECRET}" "${JWT_SUB}" "${JWT_ROLE}" "${JWT_COMPANY_ID}" "${JWT_EXP_SECONDS}" <<'PY'
import base64
import hashlib
import hmac
import json
import sys
import time

secret = sys.argv[1].encode("utf-8")
sub = sys.argv[2]
role = sys.argv[3]
company_id = sys.argv[4]
exp_seconds = int(sys.argv[5])

header = {"alg": "HS256", "typ": "JWT"}
payload = {
    "sub": sub,
    "role": role,
    "company_id": company_id,
    "exp": int(time.time()) + exp_seconds,
}

def b64(value):
    raw = json.dumps(value, separators=(",", ":"), ensure_ascii=True).encode("utf-8")
    return base64.urlsafe_b64encode(raw).rstrip(b"=").decode("utf-8")

message = f"{b64(header)}.{b64(payload)}".encode("utf-8")
signature = base64.urlsafe_b64encode(
    hmac.new(secret, message, hashlib.sha256).digest()
).rstrip(b"=").decode("utf-8")
print(f"{message.decode('utf-8')}.{signature}")
PY
