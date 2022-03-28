#!/bin/bash
set -e

# Always rollback shell options before exiting or returning
trap "set +e" EXIT RETURN

echo "-------"
echo "starting todolist-service, you can pass flags to docker compose after the profile argument normally"
echo "-------"

echo "[+] Run containers ${@} "
CERTS_DIR=./todolist/certs docker-compose -f docker-compose.dev.yml up ${@}

echo "[+] Cleaning up stopped containers..."
docker ps --all --filter status=exited -q | xargs docker rm -v;
