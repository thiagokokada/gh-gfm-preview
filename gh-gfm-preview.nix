{
  lib,
  buildGoModule,
  version ? "unknown",
}:

buildGoModule {
  pname = "gh-gfm-preview";
  inherit version;
  src = lib.cleanSource ./.;
  vendorHash = "sha256-v6NM+hihyu1/wgvcet+gfl6HLix33gCNNBsW3bO9DWc=";

  env.CGO_ENABLED = "0";

  ldflags = [
    "-s"
    "-w"
  ];

  meta = with lib; {
    description = "A Go program to preview GitHub Flavored Markdown";
    homepage = "https://github.com/thiagokokada/gh-gfm-preview";
    license = licenses.mit;
    mainProgram = "gh-gfm-preview";
  };
}
