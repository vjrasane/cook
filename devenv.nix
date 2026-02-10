{
  pkgs,
  lib,
  ...
}:
{
  dotenv.enable = true;

  env = {
    MENU_VERBOSE = "true";
  };

  packages = with pkgs; [
    cook-cli
  ];

  git-hooks.hooks = {
    nixfmt.enable = true;
    check-shebang-scripts-are-executable.enable = true;
    check-symlinks.enable = true;
    check-yaml.enable = true;
    ripsecrets.enable = true;
    shellcheck.enable = true;
    shfmt.enable = true;
    trim-trailing-whitespace.enable = true;
    end-of-file-fixer.enable = true;
  };
}
