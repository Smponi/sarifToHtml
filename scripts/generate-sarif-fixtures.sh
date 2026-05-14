#!/usr/bin/env sh
set -eu

image="${SARIF_FIXTURE_IMAGE:-sarif-fixture-tools:latest}"
target="${1:-all}"
report_dir="${REPORT_DIR:-reports}"
dockerfile="${DOCKERFILE:-tools/sarif-fixtures/Dockerfile}"
golangci_lint_version="${GOLANGCI_LINT_VERSION:-v2.12.2}"
detekt_version="${DETEKT_VERSION:-1.23.8}"
gitleaks_version="${GITLEAKS_VERSION:-v8.30.1}"
go_target="${GO_TARGET:-fixtures/scan-targets/go-bad}"
docker_target="${DOCKER_TARGET:-fixtures/scan-targets/docker-bad}"
kotlin_target="${KOTLIN_TARGET:-fixtures/scan-targets/kotlin-bad}"
semgrep_target="${SEMGREP_TARGET:-fixtures/scan-targets/semgrep-bad}"
semgrep_config="${SEMGREP_CONFIG:-fixtures/scan-targets/semgrep-rules.yml}"
build_done=0

uid="$(id -u)"
gid="$(id -g)"

mkdir -p "$report_dir"
report_path="/work/$report_dir"
cache_dir="$report_dir/.cache"
cache_path="/work/$cache_dir"

cleanup() {
  if [ "${CLEAN_IMAGE:-0}" = "1" ] && [ "$build_done" = "1" ]; then
    docker image rm "$image" >/dev/null
  fi
}

trap cleanup EXIT

docker build \
  --build-arg "GOLANGCI_LINT_VERSION=$golangci_lint_version" \
  --build-arg "DETEKT_VERSION=$detekt_version" \
  --build-arg "GITLEAKS_VERSION=$gitleaks_version" \
  -f "$dockerfile" \
  -t "$image" \
  .
build_done=1

run_in_image() {
  mkdir -p "$cache_dir/home" "$cache_dir/go-build" "$cache_dir/go-mod" "$cache_dir/golangci-lint" "$cache_dir/trivy" "$cache_dir/grype"

  docker run --rm \
    --user "${uid}:${gid}" \
    --volume "$PWD:/work" \
    --workdir /work \
    --env PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin \
    --env HOME="$cache_path/home" \
    --env XDG_CACHE_HOME="$cache_path" \
    --env GOCACHE="$cache_path/go-build" \
    --env GOMODCACHE="$cache_path/go-mod" \
    --env GOLANGCI_LINT_CACHE="$cache_path/golangci-lint" \
    --env TRIVY_CACHE_DIR="$cache_path/trivy" \
    --env GRYPE_DB_CACHE_DIR="$cache_path/grype" \
    "$image" \
    sh -c "$1"
}

golangci_lint() {
  run_in_image "mkdir -p '$report_path' && cd '$go_target' && go mod download && golangci-lint run --issues-exit-code=0 --output.sarif.path '$report_path/golangci-lint.sarif' ./..."
}

gosec() {
  run_in_image "mkdir -p '$report_path' && cd '$go_target' && go mod download && gosec -no-fail -fmt sarif -out '$report_path/gosec.sarif' ./..."
}

govulncheck() {
  run_in_image "mkdir -p '$report_path' && cd '$go_target' && go mod download && govulncheck -format sarif ./... > '$report_path/govulncheck.sarif' || test -s '$report_path/govulncheck.sarif'"
}

trivy_fs() {
  run_in_image "mkdir -p '$report_path' && trivy fs --exit-code 0 --scanners vuln,secret,misconfig,license --format sarif --output '$report_path/trivy-fs.sarif' '$docker_target'"
}

grype_fs() {
  run_in_image "mkdir -p '$report_path' && cd '$go_target' && go mod download && cd /work && grype dir:'$go_target' -o sarif > '$report_path/grype.sarif'"
}

osv_scanner() {
  run_in_image "mkdir -p '$report_path' && cd '$go_target' && go mod download && cd /work && osv-scanner scan --format sarif '$go_target' > '$report_path/osv-scanner.sarif' || test -s '$report_path/osv-scanner.sarif'"
}

gitleaks_scan() {
  run_in_image "mkdir -p '$report_path' && gitleaks dir --no-banner --redact --exit-code 0 --config fixtures/scan-targets/gitleaks.toml --report-format sarif --report-path '$report_path/gitleaks.sarif' fixtures/scan-targets"
}

semgrep_scan() {
  run_in_image "mkdir -p '$report_path' && semgrep scan --metrics=off --config '$semgrep_config' --sarif '$semgrep_target' > '$report_path/semgrep.sarif' || test -s '$report_path/semgrep.sarif'"
}

detekt_scan() {
  run_in_image "mkdir -p '$report_path' && detekt --input '$kotlin_target/src/main/kotlin' --config '$kotlin_target/detekt.yml' --base-path /work --report sarif:'$report_path/detekt.sarif' || test -s '$report_path/detekt.sarif'"
}

case "$target" in
  all)
    golangci_lint
    gosec
    # govulncheck
    trivy_fs
    grype_fs
    osv_scanner
    gitleaks_scan
    semgrep_scan
    detekt_scan
    ;;
  golangci-lint) golangci_lint ;;
  gosec) gosec ;;
  govulncheck) govulncheck ;;
  trivy|trivy-fs) trivy_fs ;;
  grype) grype_fs ;;
  osv|osv-scanner) osv_scanner ;;
  gitleaks) gitleaks_scan ;;
  semgrep) semgrep_scan ;;
  detekt) detekt_scan ;;
  *)
    echo "usage: $0 [all|golangci-lint|gosec|govulncheck|trivy|grype|osv-scanner|gitleaks|semgrep|detekt]" >&2
    exit 64
    ;;
esac

echo "SARIF reports written to $report_dir/"
