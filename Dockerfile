FROM nixos/nix
RUN nix-channel --update
COPY ./nix/nix.conf /etc/nix/nix.conf
RUN mkdir /templ
WORKDIR /templ
COPY ./flake.nix /templ/flake.nix
COPY ./flake.lock /templ/flake.lock
COPY ./nix /templ/nix
RUN nix develop --impure --command printf "Build complete\n"
COPY ./nix/.config /root/.config
CMD nix develop --impure
