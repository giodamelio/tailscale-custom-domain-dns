{
  description = "Tailscale DNS server";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  # TODO: we just need this until buildGo121Module hits nixpkgs unstable
  inputs.nixpkgs-master.url = "github:NixOS/nixpkgs/master";
  inputs.flake-parts.url = "github:hercules-ci/flake-parts";

  outputs = inputs @ { self, flake-parts, ... }: flake-parts.lib.mkFlake {inherit inputs;} {
    systems = ["x86_64-linux" "aarch64-linux"];

    perSystem = {
        pkgs,
        inputs',
        config,
        self',
        system,
        ...
      }: {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go_1_21
          ];
        };
        packages.default = inputs'.nixpkgs-master.legacyPackages.buildGo121Module {
          pname = "tailscale-custom-domain-dns";
          version = "0.6.3";
          src = ./.;
          vendorHash = null;
        };
      };
  };
}
