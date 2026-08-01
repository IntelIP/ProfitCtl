class Profitctl < Formula
  desc "CLI for profit-first unit economics simulations"
  homepage "https://github.com/IntelIP/ProfitCtl"
  version "0.1.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_darwin_arm64.tar.gz"
      sha256 "fa4734c0ef5e2111f6d2a7b024ec10d01675fd27e56c932bd14fcec61ec16999"
    else
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_darwin_amd64.tar.gz"
      sha256 "287e883308c2793fb12ff9e3c9d9c09bce69522925555e034bfdd31805295f8c"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_linux_arm64.tar.gz"
      sha256 "aeb904b5fe0b9e124a8073e6ea542852f7aefcf423fff6896965bd33f7365fd3"
    else
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_linux_amd64.tar.gz"
      sha256 "4d3ecef622d3876665c8cc20a30b4dcb8515c3abc13ebe7014945fa9172fa7f8"
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
