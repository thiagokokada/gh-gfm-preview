{
  description = "gh-gfm-preview";

  inputs = {
    nixpkgs.url = "nixpkgs/nixpkgs-unstable";
    flake-compat.url = "github:edolstra/flake-compat";
  };

  outputs =
    { self, nixpkgs, ... }:
    let
      version = "nix-${self.shortRev or self.dirtyShortRev or "unknown-dirty"}";

      supportedSystems = [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = nixpkgs.lib.getExe self.packages.${system}.default;
        };
      });

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              # ideally this would be the same version as used in CI,
              # but this works for now
              golangci-lint
            ];
            inputsFrom = [ self.packages.${system}.default ];
          };
        }
      );

      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = self.packages.${system}.gh-gfm-preview;
          gh-gfm-preview = pkgs.callPackage ./gh-gfm-preview.nix { inherit version; };
        }
      );

      formatter = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        pkgs.nixfmt-rfc-style
      );
    };
}
