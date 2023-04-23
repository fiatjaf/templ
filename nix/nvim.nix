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
in
  pkgs.neovim.override {
    vimAlias = true;
    configure = {
      packages.myPlugins = with pkgs.vimPlugins; {
       start = [
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
