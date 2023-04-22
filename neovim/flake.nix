{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "nixpkgs/nixos-22.11";
    xc = {
      url = "github:joerdav/xc";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    go = {
      url = "github:a-h/nix-golang";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, flake-utils, nixpkgs, xc, go }:
    flake-utils.lib.eachDefaultSystem (system:
    let 
      pkgs2 = import nixpkgs {};
        pkgs = import nixpkgs { 
          inherit system; overlays = [ 
            (self: super: {
              xc = xc.packages.${system}.xc;
              neovim = import ./nvim.nix { pkgs = pkgs2; };
              go = go.packages.${system}.go_1_19_2;
              gopls = pkgs.callPackage ./gopls.nix { };
              nerdfonts = (pkgs.nerdfonts.override { fonts = [ "IBMPlexMono" ]; });
            })
          ];
        };
        shell = pkgs.mkShell {
            packages = [ 
              pkgs.go
              pkgs.xc
              pkgs.neovim
              #nerdfonts
              #pkgs.asciinema
              #pkgs.git
              #pkgs.gotools
              #gopls
              #pkgs.ibm-plex
              #pkgs.ripgrep
              #pkgs.tmux
              #pkgs.wget
              #pkgs.zip
            ];
          };
      in
      {
        devShells = {
          default = shell;
        };
      }
    );
}
