#!/bin/sh
set -eu

required_vars="MONGO_BOOTSTRAP_ROOT_USERNAME MONGO_BOOTSTRAP_ROOT_PASSWORD MONGO_ADMIN_USERNAME MONGO_ADMIN_PASSWORD MONGO_HOST MONGO_PORT"

for var_name in $required_vars; do
  eval "value=\${$var_name:-}"
  if [ -z "$value" ]; then
    echo "Missing required environment variable: $var_name" >&2
    exit 1
  fi
done

until mongosh "mongodb://${MONGO_BOOTSTRAP_ROOT_USERNAME}:${MONGO_BOOTSTRAP_ROOT_PASSWORD}@${MONGO_HOST}:${MONGO_PORT}/admin?authSource=admin" --quiet --eval "db.runCommand({ ping: 1 }).ok" >/tmp/mongo-ready.out 2>/tmp/mongo-ready.err; do
  echo "Waiting for MongoDB to become ready..." >&2
  sleep 2
done

result=$(tr -d '\r\n' </tmp/mongo-ready.out)
if [ "$result" != "1" ]; then
  echo "MongoDB ping did not return success." >&2
  cat /tmp/mongo-ready.err >&2 || true
  exit 1
fi

mongosh --nodb --quiet --file /workspace/scripts/mongo/reconcile-admin.mjs
