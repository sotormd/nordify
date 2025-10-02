{
  description = "Recolor images using the Nord palette & more.";

  inputs = { nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable"; };

  outputs = { self, nixpkgs, ... }@inputs:
    let pkgs = nixpkgs.legacyPackages.x86_64-linux;
    in {
      devShells.x86_64-linux.default =
        pkgs.mkShell { packages = with pkgs; [ go gopls ]; };

      packages.x86_64-linux.default = pkgs.buildGoModule {
        pname = "nordify";
        version = "0.1.0";
        src = ./.;
        subPackages = [ "./cmd/nordify" ];
        vendorHash = null;
      };

      apps.x86_64-linux.default = {
        type = "app";
        program = "${self.packages.x86_64-linux.default}/bin/nordify";
      };
    };
}
