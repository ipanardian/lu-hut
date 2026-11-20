#!/bin/bash

set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
test_root=$(mktemp -d)
trap 'rm -rf "$test_root"' EXIT

mkdir -p "$test_root/bin" "$test_root/home/.local/bin" "$test_root/archive"
printf '#!/bin/sh\necho lu\n' > "$test_root/archive/lu"
chmod +x "$test_root/archive/lu"

os_name=$(uname -s)
arch_name=$(uname -m)
case "$arch_name" in
  x86_64) arch_name=x86_64 ;;
  arm64|aarch64) arch_name=arm64 ;;
  *) echo "unsupported test architecture: $arch_name" >&2; exit 1 ;;
esac

archive_name="lu-hut_0.7.0_${os_name}_${arch_name}.tar.gz"
tar -czf "$test_root/$archive_name" -C "$test_root/archive" lu
if command -v sha256sum >/dev/null 2>&1; then
  archive_checksum=$(sha256sum "$test_root/$archive_name" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
  archive_checksum=$(shasum -a 256 "$test_root/$archive_name" | awk '{print $1}')
else
  echo "no SHA-256 utility found (sha256sum or shasum)" >&2
  exit 1
fi

cat > "$test_root/bin/curl" <<'EOF'
#!/bin/sh
output=""
url=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o)
      output=$2
      shift 2
      ;;
    *)
      url=$1
      shift
      ;;
  esac
done

case "$url" in
  *api.github.com*)
    printf '{"tag_name":"v0.7.0"}'
    ;;
  *checksums.txt)
    case "$TEST_SCENARIO" in
      success) printf '%s  %s\n' "$TEST_ARCHIVE_CHECKSUM" "$TEST_ARCHIVE_NAME" > "$output" ;;
      missing) printf '%s  other-file.tar.gz\n' "$TEST_ARCHIVE_CHECKSUM" > "$output" ;;
      mismatch) printf '%064d  %s\n' 0 "$TEST_ARCHIVE_NAME" > "$output" ;;
    esac
    ;;
  *)
    cp "$TEST_ARCHIVE_PATH" "$output"
    ;;
esac
EOF
chmod +x "$test_root/bin/curl"

run_installer() {
  scenario=$1
  HOME="$test_root/home" \
    PATH="$test_root/bin:$PATH" \
    TEST_SCENARIO="$scenario" \
    TEST_ARCHIVE_PATH="$test_root/$archive_name" \
    TEST_ARCHIVE_NAME="$archive_name" \
    TEST_ARCHIVE_CHECKSUM="$archive_checksum" \
    bash "$repo_root/install.sh" >/dev/null 2>&1
}

run_installer success
test -x "$test_root/home/.local/bin/lu"

rm -f "$test_root/home/.local/bin/lu"
if run_installer missing; then
  echo "installer accepted a release without an archive checksum" >&2
  exit 1
fi
test ! -e "$test_root/home/.local/bin/lu"

if run_installer mismatch; then
  echo "installer accepted an archive with a mismatched checksum" >&2
  exit 1
fi
test ! -e "$test_root/home/.local/bin/lu"

echo "install checksum tests passed"
