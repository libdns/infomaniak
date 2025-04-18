FROM mcr.microsoft.com/devcontainers/go:1.24

# Install xcaddy to test directly with caddy
RUN go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest

# Install infomaniak caddy module for testing
RUN git clone https://github.com/caddy-dns/infomaniak.git /workspaces/caddy-dns-infomaniak
WORKDIR /workspaces/caddy-dns-infomaniak
RUN go mod edit -replace github.com/libdns/infomaniak=/workspaces/libdns-infomaniak

RUN mkdir /workspaces/caddy \\
    && chown -R vscode:vscode /go \\
    && chown -R vscode:vscode /workspaces \\
    && ln -s /workspaces/libdns-infomaniak/.devcontainer/.caddyfile /workspaces/caddy/.caddyfile

WORKDIR /workspaces/libdns-infomaniak

USER vscode
