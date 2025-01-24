{ pkgs ? import <nixpkgs> {} }:
  pkgs.mkShell {
    nativeBuildInputs = with pkgs.buildPackages; [
      go
      gopls
      nil
      reflex
    ];

    # shellHook = ''
    # '';
}
