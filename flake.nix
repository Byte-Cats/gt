{
  description = "GT Terminal and GF File Browser";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        goVersion = pkgs.go_1_22; # Specify Go version
        
        # Placeholder vendor hash - Needs to be generated after first build attempt
        # Run: nix build .#gt (or .#gf)
        # Copy the hash from the error message and replace fakeSha256 below.
        vendorSha256 = pkgs.lib.fakeSha256; 
        # Alternatively use: vendorSha256 = null;

      in
      {
        # Packages accessible via `nix build .#<name>`
        packages = {
          gt = pkgs.buildGoModule {
            pname = "gt";
            version = "1.2.0"; # Match version in main.go or use git describe
            src = self;
            
            subPackages = [ "cmd/gt" ];

            inherit vendorSha256;

            # System dependencies needed by gt
            buildInputs = [ 
              pkgs.sdl2
              pkgs.SDL2_ttf
            ];
            
            # On Darwin, we might need specific frameworks
            nativeBuildInputs = pkgs.lib.optionals pkgs.stdenv.isDarwin [
              pkgs.darwin.apple_sdk.frameworks.Cocoa
              pkgs.darwin.apple_sdk.frameworks.Security
            ];

            ldflags = [ "-s" "-w" ]; # Strip debug info and symbols
          };

          gf = pkgs.buildGoModule {
            pname = "gf";
            version = "0.1.0"; # Assign a version
            src = self;

            subPackages = [ "cmd/gf" ];

            inherit vendorSha256;

            # gf seems to be pure Go TUI, likely no extra system deps needed
            buildInputs = [ ]; 

            ldflags = [ "-s" "-w" ]; # Strip debug info and symbols
          };

          default = self.packages.${system}.gt; # Default package is gt
        };

        # Apps accessible via `nix run .#<name>`
        apps = {
          gt = flake-utils.lib.mkApp { drv = self.packages.${system}.gt; };
          gf = flake-utils.lib.mkApp { drv = self.packages.${system}.gf; };
          default = self.apps.${system}.gt; # Default app is gt
        };

        # Development shell accessible via `nix develop`
        devShells.default = pkgs.mkShell {
          name = "gt-dev-shell";
          
          # Tools needed for development
          packages = [
            goVersion # Go compiler
            pkgs.gopls  # Go Language Server
            pkgs.gotools # Go tools (like goimports)
            pkgs.go-outline
            pkgs.delve # Go debugger
          ];

          # Dependencies needed to build the project within the shell
          inputsFrom = [
            self.packages.${system}.gt # Includes SDL deps
            self.packages.${system}.gf
          ];

          # Environment variables if needed
          # shellHook = ''
          #   export SOME_VAR="value"
          # ''';
        };
      }
    );
} 