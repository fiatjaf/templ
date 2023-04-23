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
          (pluginGit "neovim" "nvim-lspconfig" "cf95480e876ef7699bf08a1d02aa0ae3f4d5f353" "mvDg+aT5lldqckQFpbiBsXjnwozzjKf+v3yBEyvcVLo=")
          # Configure autocomplete.
          (pluginGit "hrsh7th" "nvim-cmp" "777450fd0ae289463a14481673e26246b5e38bf2" "CoHGIiZrhRAHZ/Er0JSQMapI7jwllNF5OysLlx2QEik=")
          # Snippets manager.
          (pluginGit "hrsh7th" "vim-vsnip" "7753ba9c10429c29d25abfd11b4c60b76718c438" "ehPnvGle7YrECn76YlSY/2V7Zeq56JGlmZPlwgz2FdE=")
          (pluginGit "hrsh7th" "cmp-nvim-lsp" "0e6b2ed705ddcff9738ec4ea838141654f12eeef" "DxpcPTBlvVP88PDoTheLV2fC76EXDqS2UpM5mAfj/D4=")
        ];
        opt = [ ];
      };
      customRC = "lua require('init')";
    };
  }
