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
      pkgsDefault = import nixpkgs {};
        pkgs = import nixpkgs { 
          inherit system; overlays = [ 
            (self: super: {
              xc = xc.packages.${system}.xc;
              neovim = import ./nix/nvim.nix { pkgs = pkgsDefault; };
              go = go.packages.${system}.go_1_20_3;
              gopls = pkgs.callPackage ./nix/gopls.nix { };
              templ = pkgs.callPackage ./nix/templ.nix { 
                go = self.go; 
                xc = self.xc;
              };
              nerdfonts = (pkgsDefault.nerdfonts.override { fonts = [ "IBMPlexMono" ]; });
            })
          ];
        };
        shell = pkgs.mkShell {
            packages = [ 
              pkgs.asciinema
              pkgs.git
              pkgs.go
              pkgs.gopls
              pkgs.gotools
              pkgs.ibm-plex
              pkgs.neovim
              pkgs.nerdfonts
              pkgs.ripgrep
              pkgs.templ
              pkgs.tmux
              pkgs.wget
              pkgs.xc
              pkgs.zip
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
