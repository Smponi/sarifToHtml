ARG GO_IMAGE=golang:1.26-bookworm
FROM ${GO_IMAGE}

ARG DETEKT_VERSION=1.23.8
ARG GITLEAKS_VERSION=v8.30.1
ARG GOLANGCI_LINT_VERSION=v2.12.2

ENV GOBIN=/usr/local/bin \
    GOCACHE=/tmp/go-cache \
    GOMODCACHE=/tmp/gomod-cache \
    HOME=/tmp \
    PIPX_HOME=/opt/pipx \
    PIPX_BIN_DIR=/usr/local/bin \
    TRIVY_CACHE_DIR=/tmp/trivy-cache

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        git \
        openjdk-17-jre-headless \
        python3 \
        python3-pip \
        pipx \
        unzip \
    && rm -rf /var/lib/apt/lists/*

RUN curl -sSfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh \
        | sh -s -- -b /usr/local/bin \
    && curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh \
        | sh -s -- -b /usr/local/bin

RUN go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}" \
    && go install github.com/securego/gosec/v2/cmd/gosec@latest \
    && go install golang.org/x/vuln/cmd/govulncheck@latest \
    && go install github.com/google/osv-scanner/v2/cmd/osv-scanner@latest \
    && go install "github.com/zricethezav/gitleaks/v8@${GITLEAKS_VERSION}"

RUN pipx install semgrep

RUN curl -sSLo /tmp/detekt-cli.zip "https://github.com/detekt/detekt/releases/download/v${DETEKT_VERSION}/detekt-cli-${DETEKT_VERSION}.zip" \
    && unzip -q /tmp/detekt-cli.zip -d /opt \
    && ln -s "/opt/detekt-cli-${DETEKT_VERSION}/bin/detekt-cli" /usr/local/bin/detekt \
    && rm /tmp/detekt-cli.zip

WORKDIR /work

CMD ["bash"]
