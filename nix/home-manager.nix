{
  self,
  lib,
  ...
}: {
  config,
  pkgs,
  ...
}: let
  cfg = config.programs.circumflex;
in {
  options.programs.circumflex = {
    enable = lib.mkEnableOption "Circumflex - Hacker News in your terminal";

    package = lib.mkPackageOption self.packages.${pkgs.stdenv.hostPlatform.system} "default" {};
  };

  config = lib.mkIf cfg.enable {
    home.packages = [cfg.package];
  };
}
