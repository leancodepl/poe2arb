# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Poe2arb < Formula
  desc "POEditor JSON to Flutter ARB converter."
  homepage "https://github.com/leancodepl/poe2arb"
  version "0.4.0"

  on_macos do
    url "https://github.com/leancodepl/poe2arb/releases/download/v0.4.0/poe2arb_0.4.0_darwin_all.tar.gz"
    sha256 "bdc7219302e7593104738fa90018b0bce590e3bf2d76d41351040225aab5bdfe"

    def install
      bin.install "poe2arb"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/leancodepl/poe2arb/releases/download/v0.4.0/poe2arb_0.4.0_linux_amd64.tar.gz"
      sha256 "e9c912bb0d4dd535e0b9aeef86f020621a4941552163a16f8894f3180e122c9b"

      def install
        bin.install "poe2arb"
      end
    end
  end
end
