class Profitctl < Formula
  desc "CLI for profit-first unit economics simulations"
  homepage "https://github.com/IntelIP/ProfitCtl"
  version "0.2.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_darwin_arm64.tar.gz"
      sha256 "6476dac94d391f40ea21c4555fce1c5814fe93bb4c4b35a8ea2788dece5041c5"
    else
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_darwin_amd64.tar.gz"
      sha256 "b36d0fe67c9f869b630e3e68103ed8ad4cb49ca9f03e197fac3d4478e6fd4f3a"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_linux_arm64.tar.gz"
      sha256 "b847a1f690bac63dd4ab3b14fa47bfcc365719f6d962e785ce4f0aa27bd320f0"
    else
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_linux_amd64.tar.gz"
      sha256 "a1acd22aedf1726faba6561347d871cfe63937905b2dc714769aa0ab1aca6b00"
    end
  end

  def install
    bin.install "profitctl"
    pkgshare.install "README.md" if File.exist?("README.md")
  end

  test do
    assert_match "profit-first unit economics simulations", shell_output("#{bin}/profitctl --help")
  end
end
