{
  description = "Tailscale DNS server";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
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
        packages.default = pkgs.buildGoModule {
          pname = "tailscale-custom-domain-dns";
          version = "0.6.2";
          src = ./.;
          vendorHash = "sha256-dNTf27ef7INXjB9hkJ651aVzAY/3Ek4QjkTWWDMngrA=";
        };
      };
  };
}
