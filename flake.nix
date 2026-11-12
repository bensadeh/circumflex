{
  description = "Circumflex - Hacker News in your terminal";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs @ {
    self,
    nixpkgs,
    flake-parts,
    ...
  }:
    flake-parts.lib.mkFlake {inherit inputs;} {
      systems = ["x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin"];

      flake = {
        homeManagerModules.default = import ./nix/home-manager.nix {
          inherit self;
          lib = nixpkgs.lib;
        };

        overlays.default = final: prev: let
          gomod2nix = inputs.gomod2nix.legacyPackages.${prev.stdenv.hostPlatform.system};
        in {
          circumflex = final.callPackage ./nix/package.nix {
            inherit (gomod2nix) buildGoApplication;
          };
        };
      };

      perSystem = {
        pkgs,
        system,
        self',
        ...
      }: let
        gomod2nix = inputs.gomod2nix.legacyPackages.${system};
      in {
        packages.default = pkgs.callPackage ./nix/package.nix {
          inherit (gomod2nix) buildGoApplication;
        };

        checks = {
          package = self'.packages.default;

          formatting = pkgs.runCommand "check-formatting" {} ''
            ${pkgs.alejandra}/bin/alejandra --check ${self} > $out
          '';

          home-manager-module = let
            eval = inputs.nixpkgs.lib.evalModules {
              modules = [
                self.homeManagerModules.default
                {
                  options.home.packages = inputs.nixpkgs.lib.mkOption {
                    type = inputs.nixpkgs.lib.types.listOf inputs.nixpkgs.lib.types.package;
                    default = [];
                  };
                }
                {config.programs.circumflex.enable = true;}
              ];
              specialArgs = {inherit pkgs;};
            };
          in
            pkgs.runCommand "check-home-manager-module" {} ''
              echo "Module evaluates successfully with ${builtins.toString (builtins.length eval.config.home.packages)} packages"
              touch $out
            '';

          overlay = let
            overlayPkgs = import inputs.nixpkgs {
              localSystem = system;
              overlays = [self.overlays.default];
            };
          in
            overlayPkgs.circumflex;
        };

        formatter = pkgs.alejandra;

        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.golangci-lint
            pkgs.goreleaser
            gomod2nix.gomod2nix
          ];
        };
      };
    };
}
