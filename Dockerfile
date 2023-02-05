FROM nixos/nix:latest

RUN nix-channel --add https://nixos.org/channels/nixpkgs-unstable nixpkgs
RUN nix-channel --update
RUN nix-env -iA nixpkgs.go

RUN go work init
COPY . /src/github.com/mikerybka/reverseproxy
RUN go work use /src/github.com/mikerybka/reverseproxy

RUN go build -o /bin/reverseproxy github.com/mikerybka/reverseproxy

ENTRYPOINT ["/bin/reverseproxy"]
