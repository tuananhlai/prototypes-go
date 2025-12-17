FROM postgres:17

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Install tools to fetch and unpack sample datasets
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends curl unzip ca-certificates; \
    rm -rf /var/lib/apt/lists/*

# Download pagila (schema + data) and dvdrental (custom format) samples
RUN set -eux; \
    mkdir -p /tmp/datasets; \
    curl -L -o /tmp/datasets/pagila-schema.sql https://raw.githubusercontent.com/devrimgunduz/pagila/master/pagila-schema.sql; \
    curl -L -o /tmp/datasets/pagila-data.sql https://raw.githubusercontent.com/devrimgunduz/pagila/master/pagila-data.sql;

# Load both sample databases on first container start
RUN set -eux; \
    cat > /docker-entrypoint-initdb.d/10-load-sample-data.sh <<'EOF' && chmod +x /docker-entrypoint-initdb.d/10-load-sample-data.sh
#!/usr/bin/env bash
set -euo pipefail

# Create dedicated databases for the samples
psql --username "$POSTGRES_USER" --dbname "${POSTGRES_DB:-postgres}" <<'SQL'
CREATE DATABASE pagila;
SQL

# Import Pagila (plain SQL)
psql --username "$POSTGRES_USER" --dbname pagila -f /tmp/datasets/pagila-schema.sql
psql --username "$POSTGRES_USER" --dbname pagila -f /tmp/datasets/pagila-data.sql
EOF
