FROM nixos/nix

RUN nix-channel --update
COPY ./nix.conf /etc/nix/nix.conf
RUN mkdir /templ
WORKDIR /templ
COPY . /templ
RUN nix develop --impure --command printf "Build complete\n"
#RUN nix-build -A pythonFull '<nixpkgs>'

CMD nix develop --impure

#RUN apt-get update
#RUN apt-get install curl -y
#RUN curl -LO https://github.com/neovim/neovim/releases/latest/download/nvim.appimage
#RUN chmod u+x nvim.appimage
# ./nvim.appimage
