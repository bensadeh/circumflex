{
  lib,
  buildGoApplication,
}:
buildGoApplication {
  pname = "circumflex";
  version = "4.4-dev";
  src = lib.cleanSourceWith {
    src = ./..;
    filter = path: type:
      (lib.hasSuffix ".go" path)
      || (lib.hasSuffix ".mod" path)
      || (type == "directory");
  };
  modules = ../gomod2nix.toml;
  doCheck = false;

  meta = {
    description = "It's Hacker News in your terminal";
    homepage = "https://github.com/bensadeh/circumflex";
    license = lib.licenses.mit;
    mainProgram = "circumflex";
  };
}
