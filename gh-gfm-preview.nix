{
  lib,
  buildGoModule,
  version ? "unknown",
}:

buildGoModule {
  pname = "gh-gfm-preview";
  inherit version;
  src = lib.cleanSource ./.;
  vendorHash = "sha256-NF0ArGe34wNHUveE673EWsQ/N15TUiD0r75W4ocxkf0=";

  env.CGO_ENABLED = "0";

  ldflags = [
    "-X=main.version=${version}"
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
