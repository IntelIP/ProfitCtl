import { expect, test } from "bun:test";
import { nativeTarget } from "../lib/platform.js";

test("maps supported Bun platforms to Go release targets", () => {
  expect(nativeTarget("darwin", "arm64")).toEqual({
    directory: "darwin-arm64",
    executable: "profitctl",
  });
  expect(nativeTarget("linux", "x64")).toEqual({
    directory: "linux-amd64",
    executable: "profitctl",
  });
  expect(nativeTarget("win32", "x64")).toEqual({
    directory: "windows-amd64",
    executable: "profitctl.exe",
  });
});

test("rejects release targets ProfitCtl does not build", () => {
  expect(() => nativeTarget("win32", "arm64")).toThrow("does not provide a binary");
  expect(() => nativeTarget("freebsd", "x64")).toThrow("does not provide a binary");
});
