# ----- EIB Builder Image -----
FROM registry.suse.com/bci/golang:1.22

ENV HAULER_VERSION=1.0.7
ENV PLATFORM=linux
ENV ARCH=amd64
ENV CGO_ENABLED=1
ENV GO111MODULE=on

# Dependency uses by line
# 1. Podman Go library
RUN zypper install -y \
    gpgme-devel device-mapper-devel libbtrfs-devel

WORKDIR /src

RUN curl -sfL "https://github.com/hauler-dev/hauler/releases/download/v${HAULER_VERSION}/hauler_${HAULER_VERSION}_${PLATFORM}_${ARCH}.tar.gz" | tar -xzf -

COPY go.mod go.sum ./
COPY ./cmd ./cmd
COPY ./pkg ./pkg
COPY .git .git

RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    --mount=type=cache,id=gobuild,target=/root/.cache/go-build \
    go build ./cmd/eib

# CGO_ENABLED=0
# ----- Deliverable Image -----
FROM opensuse/leap:15.6

# Dependency uses by line
# 1. ISO image building
# 2. RAW image modification on x86_64
# 3. Podman EIB library
# 4. RPM resolution logic
# 5. Embedded artefact registry
# 6. Network configuration
RUN zypper addrepo https://download.opensuse.org/repositories/isv:SUSE:Edge:EdgeImageBuilder/SLE-15-SP5/isv:SUSE:Edge:EdgeImageBuilder.repo && \
    zypper --gpg-auto-import-keys refresh && \
    zypper install -y \
    xorriso squashfs  \
    libguestfs kernel-default e2fsprogs parted gptfdisk btrfsprogs guestfs-tools lvm2\
    podman \
    createrepo_c \
    helm hauler \
    nm-configurator && \
    zypper clean -a

COPY --from=0 /src/eib /bin/eib
COPY --from=0 /src/hauler /opt/hauler/hauler

ENTRYPOINT ["/bin/eib"]
