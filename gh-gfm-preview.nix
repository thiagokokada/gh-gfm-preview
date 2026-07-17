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

  vendorHash = "sha256-ppp84qMK+DqmD2I9i52Op8gpP/TWmhS6Ck80MIicNUQ=";

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
