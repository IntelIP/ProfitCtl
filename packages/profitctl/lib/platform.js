const platformNames = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const architectureNames = {
  arm64: "arm64",
  x64: "amd64",
};

export function nativeTarget(platform = process.platform, architecture = process.arch) {
  const goos = platformNames[platform];
  const goarch = architectureNames[architecture];

  if (!goos || !goarch || (goos === "windows" && goarch !== "amd64")) {
    throw new Error(`ProfitCtl does not provide a binary for ${platform}/${architecture}`);
  }

  return {
    directory: `${goos}-${goarch}`,
    executable: goos === "windows" ? "profitctl.exe" : "profitctl",
  };
}
