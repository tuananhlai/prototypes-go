# Postgres database with 3 sample databases of increasing size (pagila, employees and postgres_air).
FROM postgres:17

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Install tools to fetch and unpack sample datasets
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends curl pipx ca-certificates; \
    rm -rf /var/lib/apt/lists/*

# Install Google Drive downloader CLI tool to retrieve postgres_air database later.
RUN pipx install gdown;

# Download pagila (schema + data) and dvdrental (custom format) samples
RUN set -eux; \
    mkdir -p /tmp/datasets; \
    curl -L -o /tmp/datasets/pagila-schema.sql https://raw.githubusercontent.com/devrimgunduz/pagila/master/pagila-schema.sql; \
    curl -L -o /tmp/datasets/pagila-data.sql https://raw.githubusercontent.com/devrimgunduz/pagila/master/pagila-data.sql; \
    curl -L -o /tmp/datasets/employees.sql.gz  https://raw.githubusercontent.com/neondatabase-labs/postgres-sample-dbs/main/employees.sql.gz; \
    /root/.local/bin/gdown https://drive.google.com/uc?id=1nNXcTjhJK6EzJdzaZvdoJDrYj_V61rq8 -O /tmp/datasets/postgres_air_2024.backup;

# Load both sample databases on first container start
RUN set -eux; \
    cat > /docker-entrypoint-initdb.d/10-load-sample-data.sh <<'EOF' && chmod +x /docker-entrypoint-initdb.d/10-load-sample-data.sh
#!/usr/bin/env bash
set -euo pipefail

# Create dedicated databases for the samples
psql --username "$POSTGRES_USER" --dbname "${POSTGRES_DB:-postgres}" <<'SQL'
CREATE DATABASE pagila;

CREATE DATABASE employees;

CREATE DATABASE postgres_air;
SQL

# Import Pagila (plain SQL)
psql --username "$POSTGRES_USER" --dbname pagila -f /tmp/datasets/pagila-schema.sql
psql --username "$POSTGRES_USER" --dbname pagila -f /tmp/datasets/pagila-data.sql

# Import `employees` database
pg_restore -U "$POSTGRES_USER" -d employees -Fc /tmp/datasets/employees.sql.gz -v --no-owner --no-privileges

pg_restore -U "$POSTGRES_USER" -d postgres_air -Fc /tmp/datasets/postgres_air_2024.backup -v --no-owner --no-privileges
EOF
