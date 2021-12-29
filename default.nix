{

    pkgs ? import (fetchTarball {
        name = "nixpkgs-21-11";
        url = "https://github.com/NixOS/nixpkgs/archive/a7ecde854aee5c4c7cd6177f54a99d2c1ff28a31.tar.gz";
        sha256 = "162dywda2dvfj1248afxc45kcrg83appjd0nmdb541hl7rnncf02";
    }) { },

}: rec {

  # https://go.dev/dl/#go1.18beta1
  go = fetchTarball {
    name = "go1.18beta1";
    url = "https://go.dev/dl/go1.18beta1.linux-amd64.tar.gz";
    sha256 = "09sb0viv1ybx6adgx4jym1sckdq3mpjkd6albj06hwnchj5rqn40";
  };

  shell = pkgs.mkShell {
    name = "ginger-dev";
    buildInputs = [
      go
    ];
  };

}
