#!/usr/bin/env sh
set -eu

mobilelab_bin=${MOBILELAB_BIN:-../../bin/mobilelab}
artifact_dir=${MOBILELAB_ARTIFACT_DIR:-artifacts}

mkdir -p "$artifact_dir"

"$mobilelab_bin" start --headless >"$artifact_dir/mobilelab.log" 2>&1 &
core_pid=$!

cleanup() {
  "$mobilelab_bin" stop >/dev/null 2>&1 || true
  wait "$core_pid" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

attempt=0
until "$mobilelab_bin" status >/dev/null 2>&1; do
  attempt=$((attempt + 1))
  if [ "$attempt" -ge 30 ]; then
    echo "MobileLab did not become ready; see $artifact_dir/mobilelab.log" >&2
    exit 1
  fi
  sleep 1
done

run_report() {
  format=$1
  output=$2
  "$mobilelab_bin" run scenarios --platform fake --timeout 10s --report "$format" --output "$output" &
  run_pid=$!
  sleep 1
  curl --fail --silent --show-error http://127.0.0.1:4566/health >"$artifact_dir/health-response.json"
  wait "$run_pid"
}

run_report junit "$artifact_dir/mobilelab-junit.xml"
run_report html "$artifact_dir/mobilelab-report.html"
