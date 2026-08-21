#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== Building check_ftp2 ==="
go build -o check_ftp2

echo "=== Generating test TLS certificates ==="
mkdir -p test-certs
openssl req -new -newkey rsa:2048 -days 1 -nodes -x509 \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1" \
  -keyout test-certs/key.pem \
  -out test-certs/cert.pem 2>/dev/null || true

# pure-ftpd は証明書と秘密鍵が結合された 1 つの PEM ファイルを必要とする
cat test-certs/cert.pem test-certs/key.pem > test-certs/pureftpd.pem
chmod 600 test-certs/pureftpd.pem

# crazymax/pure-ftpd 用の追加フラグを設定
cat > test-certs/pureftpd.flags <<'EOF'
--tls=1
EOF

echo "=== Creating FTP user for TLS server ==="

start_containers() {
  echo "=== Starting Docker Compose test environment ==="
  docker compose -f docker-compose.test.yml down -v 2>/dev/null || true
  docker compose -f docker-compose.test.yml up -d
  # TLS サーバー起動後にユーザーを作成
  docker compose -f docker-compose.test.yml exec -T ftp-tls \
    sh -c 'pure-pw useradd test -u 1000 -g 1000 -d /home/test -m <<PW
test
test
PW'
}

start_containers

wait_for_ftp() {
  local host="$1"
  local port="$2"
  local retries=30
  for i in $(seq 1 "$retries"); do
    if nc -z "$host" "$port" 2>/dev/null; then
      return 0
    fi
    echo "Waiting for $host:$port... ($i/$retries)"
    sleep 1
  done
  echo "Timeout waiting for $host:$port"
  return 1
}

cleanup() {
  echo "=== Stopping Docker Compose test environment ==="
  docker compose -f docker-compose.test.yml down -v 2>/dev/null || true
  rm -rf test-certs
}
trap cleanup EXIT

wait_for_ftp localhost 2121
wait_for_ftp localhost 2122

echo "=== Testing plain FTP ==="
./check_ftp2 -H localhost -p 2121

echo "=== Testing explicit TLS without verification ==="
./check_ftp2 -H localhost -p 2122 --explicit

echo "=== Testing explicit TLS with certificate verification ==="
./check_ftp2 -H localhost -p 2122 --explicit --verify-ssl --sni localhost

echo "=== All integration tests passed ==="
