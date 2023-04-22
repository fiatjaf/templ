{ pkgs, ... }:

let
  pluginGit = owner: repo: ref: sha: pkgs.vimUtils.buildVimPluginFrom2Nix {
    pname = "${repo}";
    version = ref;
    src = pkgs.fetchFromGitHub {
      owner = owner;
      repo = repo;
      rev = ref;
      sha256 = sha;
    };
  };
  # See list of grammars.
  # https://github.com/NixOS/nixpkgs/blob/c4fcf9a2cc4abde8a2525691962b2df6e7327bd3/pkgs/applications/editors/vim/plugins/nvim-treesitter/generated.nix
  nvim-treesitter-with-plugins = pkgs.vimPlugins.nvim-treesitter.withPlugins (p: [ 
    p.go
    p.gomod
    p.javascript
    p.json
    p.nix
    p.html
    p.tsx
    p.typescript
    p.yaml
    p.dockerfile
  ]);
in
  pkgs.neovim.override {
    vimAlias = true;
    configure = {
      packages.myPlugins = with pkgs.vimPlugins; {
       start = [
          (pluginGit "Mofiqul" "dracula.nvim" "a0b129d7dea51b317fa8064f13b29f68004839c4" "snCRLw/QtKPDAkh1CXZfto2iCoyaQIx++kOEC0vy9GA=")
          # Tressiter syntax highlighting.
          nvim-treesitter-with-plugins
          # Templ highlighting.
          (pluginGit "Joe-Davidson1802" "templ.vim" "2d1ca014c360a46aade54fc9b94f065f1deb501a" "1bc3p0i3jsv7cbhrsxffnmf9j3zxzg6gz694bzb5d3jir2fysn4h")
          # Add function signatures to autocomplete.
          (pluginGit "ray-x" "lsp_signature.nvim" "1fdc742af68f4725a22c05c132f971143be447fc" "DITo8Sp/mcOPhCTcstmpb1i+mUc5Ao8PEP5KYBO8Xdg=")
          # Configure autocomplete.
          (pluginGit "hrsh7th" "nvim-cmp" "777450fd0ae289463a14481673e26246b5e38bf2" "CoHGIiZrhRAHZ/Er0JSQMapI7jwllNF5OysLlx2QEik=")
          # Configure autocomplete.
          (pluginGit "neovim" "nvim-lspconfig" "cf95480e876ef7699bf08a1d02aa0ae3f4d5f353" "mvDg+aT5lldqckQFpbiBsXjnwozzjKf+v3yBEyvcVLo=")
          # Snippets manager.
          (pluginGit "L3MON4D3" "LuaSnip" "8d6c0a93dec34900577ba725e91c44b8d3ca1f45" "ymH5iiMcCAfpU2W/2W6cFbQgZUSRSieUmdsWoizTaZg=")
          # Add snippets to the autocomplete.
          (pluginGit "saadparwaiz1" "cmp_luasnip" "18095520391186d634a0045dacaa346291096566" "Z5SPy3j2oHFxJ7bK8DP8Q/oRyLEMlnWyIfDaQcNVIS0=")
          (pluginGit "hrsh7th" "cmp-nvim-lsp" "0e6b2ed705ddcff9738ec4ea838141654f12eeef" "DxpcPTBlvVP88PDoTheLV2fC76EXDqS2UpM5mAfj/D4=")
        ];
        opt = [ ];
      };
      customRC = "lua require('init')";
    };
  }
