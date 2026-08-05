class Profitctl < Formula
  desc "CLI for profit-first unit economics simulations"
  homepage "https://github.com/IntelIP/ProfitCtl"
  version "0.3.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_darwin_arm64.tar.gz"
      sha256 "de05841ba1499cfcce8ff873d212d8077ee5c078899d2c6ee3976e79fc3a0878"
    else
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_darwin_amd64.tar.gz"
      sha256 "0b3a5c83a68efd06efbfd249b34d970c23eab65f53ee9b29ea8a488c7fc688a4"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_linux_arm64.tar.gz"
      sha256 "14d7a6f432ea28256f3f6f7916e87199a293f99ae3c93ae951f1384bf8726e12"
    else
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_linux_amd64.tar.gz"
      sha256 "cffd2fdf888198891e476109871f80300772f2f84c7654f9c9a1ec2d5049c71f"
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
