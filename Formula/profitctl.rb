class Profitctl < Formula
  desc "CLI for profit-first unit economics simulations"
  homepage "https://github.com/IntelIP/ProfitCtl"
  version "0.1.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_darwin_arm64.tar.gz"
      sha256 "7a16e98a82d96b8c1256deea599b10ee469bc5fa5e19b35a3ece803bc0b93a7d"
    else
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_darwin_amd64.tar.gz"
      sha256 "fc13529c0e44a7401d9c50ab75a56fbc63cbca0d4031bd69b737499ca264c951"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_linux_arm64.tar.gz"
      sha256 "4a744754afdbea618e07d74463d788e3e91c1fc67e506479263af8b09c7203af"
    else
      url "https://github.com/IntelIP/ProfitCtl/releases/download/v#{version}/profitctl_v#{version}_linux_amd64.tar.gz"
      sha256 "425003e38aab4133a3cc37275a97ca31a008d9a011bfbbb9425bf43dbb33c712"
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
