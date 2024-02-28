{
  inputs = {
    nixpkgs.url = "nixpkgs";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }:
    utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        fyne-cross = pkgs.buildGoModule {
          name = "fyne-cross";
          doCheck = false;
          src = pkgs.fetchFromGitHub {
            owner = "fyne-io";
            repo = "fyne-cross";
            rev = "f8b06cf33396ad174fa7169216560a9d63e6cd73";
            hash = "sha256-dNd6D85GRgjHHS1D7IVvsznC8hn8sREQU71RhshZLbY=";
          };
          vendorHash = "sha256-68WD7kij+au0cwnnbB1U1OA3Bb8vjkdPbhV6aiQcjgo=";
        };
        buildScript = pkgs.writeScriptBin "compile-windows" ''
          if [ ! -f Icon.png ]; then
            echo "Cannot find Icon.png, did you run this in the right directory?"
            exit 1
          fi

          exec ${fyne-cross}/bin/fyne-cross windows --app-id Rooster.Fixer --icon Icon.png
        '';
      in {
        devShell = pkgs.mkShell {
          name = "devshell";
          buildInputs = with pkgs; [
            fyne-cross
            buildScript
            libGL
            pkg-config
            xorg.libX11.dev
            xorg.libXcursor
            xorg.libXi
            xorg.libXinerama
            xorg.libXrandr
            xorg.libXxf86vm
            # libxkbcommon
            # wayland
            libxkbcommon
            wayland
          ];
        };
      });
}
