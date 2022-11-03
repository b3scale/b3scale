{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    postgresql
    go
    goreleaser
  ]; 
}
