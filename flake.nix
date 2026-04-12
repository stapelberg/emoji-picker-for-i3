{
  description = "Emoji picker for i3 using rofi";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-25.11";
    nixpkgs-unstable.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    stapelbergnix = {
      url = "github:stapelberg/nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      nixpkgs-unstable,
      stapelbergnix,
    }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs-unstable = import nixpkgs-unstable {
            inherit system;
            overlays = [ stapelbergnix.overlays.goVcsStamping ];
          };
        in
        {
          default = pkgs-unstable.buildGoLatestModule {
            pname = "emoji-picker-for-i3";
            version = "unstable";
            src = self;
            nativeBuildInputs = [ pkgs-unstable.makeWrapper ];
            subPackages = [ "cmd/emoji-picker-for-i3" ];
            env.CGO_ENABLED = "0";
            vendorHash = "sha256-5DilR7gmBrNlxwrKKPe7VjngxkbmR/KWGO6NW40SOms=";
            doCheck = false;

            postInstall = ''
              wrapProgram $out/bin/emoji-picker-for-i3 \
                --prefix PATH : ${
                  pkgs-unstable.lib.makeBinPath [
                    pkgs-unstable.rofi
                    pkgs-unstable.xdotool
                  ]
                }
            '';
          };
        }
      );

      nixosModules.default =
        {
          config,
          lib,
          pkgs,
          ...
        }:
        let
          cfg = config.programs.emoji-picker-for-i3;
        in
        {
          options.programs.emoji-picker-for-i3 = {
            enable = lib.mkEnableOption "emoji-picker-for-i3 emoji picker";

            package = lib.mkOption {
              type = lib.types.package;
              default = self.packages.${pkgs.system}.default;
              description = "The emoji-picker-for-i3 package to use.";
            };
          };

          config = lib.mkIf cfg.enable {
            environment.systemPackages = [ cfg.package ];
          };
        };

      formatter = forAllSystems (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        pkgs.nixfmt-tree
      );
    };
}
