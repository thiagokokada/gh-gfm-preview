{
  lib,
  buildGoModule,
  version ? "unknown",
}:

buildGoModule {
  pname = "gfm-markdown-preview";
  inherit version;
  src = lib.cleanSource ./.;
  vendorHash = "sha256-RYEaK9/CeKlPdfPogGBpRM4FgFS6ZCFJnC2UOn7x7fg=";

  env.CGO_ENABLED = "0";

  ldflags =
    [
      "-X=main.version=${version}"
      "-s"
      "-w"
    ];

  meta = with lib; {
    description = "A Go program to preview GitHub Flavored Markdown";
    homepage = "https://github.com/thiagokokada/gfm-markdown-preview";
    license = licenses.mit;
    mainProgram = "gfm-markdown-preview";
  };
}
