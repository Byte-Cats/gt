# Nix Build and Development Guide for GT/GF

This document provides instructions for building and developing the GT Terminal (`gt`) and GF File Browser (`gf`) using [Nix](https://nixos.org/) with Flakes.

This guide is intended for contributors or users who want to build the project from source using the Nix package manager.

## Prerequisites

1.  **Nix Installed**: You need to have Nix installed on your system. Follow the instructions on the [official Nix website](https://nixos.org/download.html).
2.  **Flakes Enabled**: Ensure that Nix flakes support is enabled. This usually involves adding `experimental-features = nix-command flakes` to your Nix configuration file (`/etc/nix/nix.conf` or `~/.config/nix/nix.conf`) and potentially restarting the Nix daemon.

## Initial Setup: Obtaining the Vendor Hash

The `flake.nix` file uses a fixed-output derivation (`buildGoModule`) which requires a hash of the vendored Go dependencies (`vendorSha256`). For security and reproducibility, Nix requires this hash to be specified explicitly.

The `flake.nix` initially contains a placeholder hash:

```nix
vendorSha256 = pkgs.lib.fakeSha256;
```

To obtain the correct hash, follow these steps:

1.  **Navigate to Project Root**: Open your terminal and change into the root directory where you cloned or forked this repository.
2.  **Attempt First Build**: Run the build command for either package:
    ```bash
    nix build .#gt 
    # OR
    nix build .#gf
    ```
3.  **Copy the Hash**: This command **will fail** the first time with an error message indicating the expected hash. Look for output similar to this:
    ```
    error: hash mismatch in fixed-output derivation '/nix/store/...-go-modules':
             specified: sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
                got:    sha256-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX=
    ```
    Copy the hash listed after `got:`. It will start with `sha256-`.
4.  **Update `flake.nix`**: Open the `flake.nix` file in the project root and replace `pkgs.lib.fakeSha256` with the actual hash you copied, enclosed in quotes:
    ```nix
    # vendorSha256 = pkgs.lib.fakeSha256; 
    vendorSha256 = "sha256-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX=";
    ```
5.  **Rebuild**: Run the build command (`nix build .#gt` or `nix build .#gf`) again. It should now succeed.

*(Note: If you prefer, you can initially set `vendorSha256 = null;` instead of using `fakeSha256`. The build will still fail but might provide the hash more directly in some Nix versions.)*

## Building the Applications

Once the vendor hash is correctly set, you can build the applications from the project root directory:

*   **Build GT Terminal**: 
    ```bash
    nix build .#gt
    ```
    The resulting binary will be symlinked at `./result/bin/gt`.

*   **Build GF File Browser**:
    ```bash
    nix build .#gf
    ```
    The resulting binary will be symlinked at `./result/bin/gf`.

## Running the Applications

You can run the applications directly from the project root without building them first:

*   **Run GT Terminal**: 
    ```bash
    nix run .#gt -- [arguments-for-gt]
    ```

*   **Run GF File Browser**:
    ```bash
    nix run .#gf -- [arguments-for-gf]
    ```

## Development Environment

A development shell is provided with Go, gopls, SDL dependencies, and other common Go development tools.

*   **Enter the Shell** (from the project root):
    ```bash
    nix develop
    ```
*   **Inside the Shell**: You can use standard Go commands (`go build ./cmd/gt`, `go run ./cmd/gf/main.go`, `gopls`, etc.). All necessary dependencies defined in the flake (including SDL for GT) will be available in the environment.

## Updating Dependencies & Contributing

If you modify the Go dependencies (i.e., changes to `go.mod` or `go.sum`) while working on a feature or fix:

1.  The `vendorSha256` in `flake.nix` will become invalid.
2.  You must repeat the process described in the "Initial Setup" section to obtain and update the vendor hash **before** committing your changes.
3.  Include the updated `flake.nix` with the correct `vendorSha256` in your Pull Request so that others can build your changes using Nix. 