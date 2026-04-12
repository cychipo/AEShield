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

wait_for_mongo() {
  attempt=0
  until mongosh --host "$MONGO_HOST" --port "$MONGO_PORT" --quiet --eval "db.runCommand({ ping: 1 }).ok" >/tmp/mongo-ready.out 2>/tmp/mongo-ready.err; do
    attempt=$((attempt + 1))
    if [ "$attempt" -ge 30 ]; then
      echo "MongoDB did not become reachable in time." >&2
      cat /tmp/mongo-ready.err >&2 || true
      exit 1
    fi
    echo "Waiting for MongoDB to become ready..." >&2
    sleep 2
  done

  result=$(tr -d '\r\n' </tmp/mongo-ready.out)
  if [ "$result" != "1" ]; then
    echo "MongoDB ping did not return success." >&2
    cat /tmp/mongo-ready.err >&2 || true
    exit 1
  fi
}

can_auth_with_uri() {
  uri="$1"
  mongosh "$uri" --quiet --eval "db.runCommand({ connectionStatus: 1 }).ok" >/tmp/mongo-auth.out 2>/tmp/mongo-auth.err || return 1
  result=$(tr -d '\r\n' </tmp/mongo-auth.out)
  [ "$result" = "1" ]
}

wait_for_mongo

bootstrap_uri="mongodb://${MONGO_BOOTSTRAP_ROOT_USERNAME}:${MONGO_BOOTSTRAP_ROOT_PASSWORD}@${MONGO_HOST}:${MONGO_PORT}/admin?authSource=admin"
desired_uri="mongodb://${MONGO_ADMIN_USERNAME}:${MONGO_ADMIN_PASSWORD}@${MONGO_HOST}:${MONGO_PORT}/admin?authSource=admin"

if can_auth_with_uri "$bootstrap_uri"; then
  export MONGO_RECONCILE_AUTH_MODE=bootstrap-root
elif can_auth_with_uri "$desired_uri"; then
  export MONGO_RECONCILE_AUTH_MODE=deploy-admin
else
  echo "Neither bootstrap root credentials nor deploy admin credentials can authenticate to MongoDB." >&2
  echo "If this is an existing volume without the bootstrap root user, either restore valid credentials or reset the MongoDB volume." >&2
  cat /tmp/mongo-auth.err >&2 || true
  exit 1
fi

mongosh --nodb --quiet --file /workspace/scripts/mongo/reconcile-admin.mjs
