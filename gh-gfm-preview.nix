{
  lib,
  buildGoModule,
  version ? "unknown",
}:

buildGoModule {
  pname = "gh-gfm-preview";
  inherit version;
  src = lib.cleanSource ./.;
  vendorHash = "sha256-XDTBpXRogpmBQXmKD1MHBZl5XPsRdN5ZdYhoxurC5Ws=";

  env.CGO_ENABLED = "0";

  ldflags =
    [
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
