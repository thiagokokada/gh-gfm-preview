{
  lib,
  buildGoModule,
  version ? "unknown",
}:

buildGoModule {
  pname = "gh-gfm-preview";
  inherit version;

  src = lib.fileset.toSource {
    root = ./.;
    fileset = lib.fileset.unions [
      ./cmd
      ./go.mod
      ./go.sum
      ./internal
      ./main.go
      ./testdata
    ];
  };

  env.CGO_ENABLED = "0";

  vendorHash = "sha256-xrLG+Jkm2prSG9fcnJSkWGFpxMpynYVchl9SVyxC280=";

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
