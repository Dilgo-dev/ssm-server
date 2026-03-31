#!/bin/sh
set -e

# ssm-sync installer
# Usage: curl -fsSL https://raw.githubusercontent.com/Dilgo-dev/ssm-sync/main/install.sh | sh

IMAGE="ghcr.io/dilgo-dev/ssm-sync:latest"

printf "\n  \033[1;35mssm-sync\033[0m installer\n\n"

# --- Check Docker ---
if ! command -v docker >/dev/null 2>&1; then
    printf "  \033[1;31mError:\033[0m Docker is not installed.\n"
    printf "  Install it: https://docs.docker.com/get-docker/\n\n"
    exit 1
fi

COMPOSE_CMD=""
if docker compose version >/dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
    COMPOSE_CMD="docker-compose"
else
    printf "  \033[1;31mError:\033[0m Docker Compose is not installed.\n"
    printf "  Install it: https://docs.docker.com/compose/install/\n\n"
    exit 1
fi

# --- Install directory ---
if [ "$(id -u)" = "0" ]; then
    INSTALL_DIR="/opt/ssm-sync"
else
    INSTALL_DIR="$HOME/.ssm-sync"
fi

printf "  Install directory: \033[1m%s\033[0m\n\n" "$INSTALL_DIR"

# --- Configuration ---
printf "  \033[1;35mConfiguration\033[0m\n\n"

printf "  Port [8080]: "
read -r PORT </dev/tty
PORT="${PORT:-8080}"

printf "  JWT Secret (leave empty to auto-generate): "
read -r JWT_SECRET </dev/tty
if [ -z "$JWT_SECRET" ]; then
    JWT_SECRET=$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n')
    printf "  \033[2mGenerated: %s\033[0m\n" "$JWT_SECRET"
fi

printf "\n  Configure SMTP for email verification? [y/N]: "
read -r SMTP_ANSWER </dev/tty
SMTP_ANSWER=$(echo "$SMTP_ANSWER" | tr '[:upper:]' '[:lower:]')

SMTP_HOST=""
SMTP_PORT="587"
SMTP_USER=""
SMTP_PASS=""
API_URL=""

if [ "$SMTP_ANSWER" = "y" ] || [ "$SMTP_ANSWER" = "yes" ]; then
    printf "  SMTP Host: "
    read -r SMTP_HOST </dev/tty
    printf "  SMTP Port [587]: "
    read -r SMTP_PORT_INPUT </dev/tty
    SMTP_PORT="${SMTP_PORT_INPUT:-587}"
    printf "  SMTP User: "
    read -r SMTP_USER </dev/tty
    printf "  SMTP Password: "
    read -r SMTP_PASS </dev/tty
    printf "  Public API URL (for email links, e.g. https://sync.example.com): "
    read -r API_URL </dev/tty
fi

# --- Generate files ---
printf "\n  \033[1;35mSetting up...\033[0m\n\n"

mkdir -p "$INSTALL_DIR"

cat > "$INSTALL_DIR/.env" <<EOF
JWT_SECRET=$JWT_SECRET
PORT=$PORT
DATA_DIR=/data
SMTP_HOST=$SMTP_HOST
SMTP_PORT=$SMTP_PORT
SMTP_USER=$SMTP_USER
SMTP_PASS=$SMTP_PASS
API_URL=$API_URL
EOF

cat > "$INSTALL_DIR/docker-compose.yml" <<EOF
services:
  ssm-sync:
    image: $IMAGE
    ports:
      - "$PORT:$PORT"
    env_file: .env
    volumes:
      - ssm-data:/data
    restart: unless-stopped

volumes:
  ssm-data:
EOF

# --- Start ---
cd "$INSTALL_DIR"
$COMPOSE_CMD pull -q
$COMPOSE_CMD up -d

# --- Verify ---
printf "\n  Waiting for server..."
TRIES=0
while [ $TRIES -lt 10 ]; do
    if curl -s "http://localhost:$PORT/health" | grep -q '"ok"' 2>/dev/null; then
        break
    fi
    sleep 1
    TRIES=$((TRIES + 1))
done

if curl -s "http://localhost:$PORT/health" | grep -q '"ok"' 2>/dev/null; then
    printf " \033[1;32mrunning!\033[0m\n"
else
    printf " \033[1;31mfailed to start\033[0m\n"
    printf "  Check logs: cd %s && %s logs\n\n" "$INSTALL_DIR" "$COMPOSE_CMD"
    exit 1
fi

# --- Summary ---
printf "\n  \033[1;35mssm-sync is ready\033[0m\n\n"
printf "  Server:    \033[1mhttp://localhost:%s\033[0m\n" "$PORT"
printf "  Config:    %s/.env\n" "$INSTALL_DIR"
printf "  Data:      docker volume (ssm-data)\n"
printf "  Logs:      cd %s && %s logs -f\n" "$INSTALL_DIR" "$COMPOSE_CMD"
printf "  Stop:      cd %s && %s down\n" "$INSTALL_DIR" "$COMPOSE_CMD"
printf "  Update:    cd %s && %s pull && %s up -d\n\n" "$INSTALL_DIR" "$COMPOSE_CMD" "$COMPOSE_CMD"
printf "  In ssm, set the server to \033[1mhttp://YOUR_IP:%s\033[0m\n" "$PORT"
printf "  during \033[1mssm register\033[0m or \033[1mssm login\033[0m.\n\n"
