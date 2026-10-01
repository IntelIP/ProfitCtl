type Callback = (error: Error | null, value?: unknown) => void;

// Only the bundled fictional scenarios exist in this read-only browser workspace.
export function createDemoFilesystem(files: Record<string, string>, output: (text: string) => void) {
  const encoder = new TextEncoder();
  const decoder = new TextDecoder();
  const contents = new Map(Object.entries(files).map(([name, text]) => [`/${name}`, encoder.encode(text)]));
  const descriptors = new Map<number, { bytes: Uint8Array; offset: number }>();
  let nextDescriptor = 3;
  let written = 0;
  const error = (code: string) => Object.assign(new Error(code), { code });
  const done = (callback: Callback, failure: Error | null, value?: unknown) => queueMicrotask(() => callback(failure, value));
  const lookup = (name: string) => name.includes("\0") || name.split("/").includes("..") ? undefined : contents.get(name.startsWith("/") ? name : `/${name}`);
  const stat = (size: number) => ({ dev: 0, ino: 1, mode: 0o100444, nlink: 1, uid: 0, gid: 0, rdev: 0, size, blksize: 4096, blocks: Math.ceil(size / 512), atimeMs: 0, mtimeMs: 0, ctimeMs: 0, isDirectory: () => false });
  const filesystem = {
    constants: { O_WRONLY: 1, O_RDWR: 2, O_CREAT: 64, O_TRUNC: 512, O_APPEND: 1024, O_EXCL: 128, O_DIRECTORY: 65536 },
    open(name: string, flags: number, _mode: number, callback: Callback) {
      const bytes = lookup(name);
      if (flags !== 0) return done(callback, error("EROFS"));
      if (!bytes) return done(callback, error("ENOENT"));
      const descriptor = nextDescriptor++;
      descriptors.set(descriptor, { bytes, offset: 0 });
      done(callback, null, descriptor);
    },
    read(descriptor: number, buffer: Uint8Array, offset: number, length: number, position: number | null, callback: Callback) {
      if (descriptor === 0) return done(callback, null, 0);
      const file = descriptors.get(descriptor);
      if (!file) return done(callback, error("EBADF"));
      const start = position ?? file.offset;
      const bytes = file.bytes.subarray(start, start + length);
      buffer.set(bytes, offset);
      file.offset = start + bytes.length;
      done(callback, null, bytes.length);
    },
    close(descriptor: number, callback: Callback) {
      done(callback, descriptors.delete(descriptor) ? null : error("EBADF"));
    },
    fstat(descriptor: number, callback: Callback) {
      const file = descriptors.get(descriptor);
      done(callback, file || descriptor < 3 ? null : error("EBADF"), stat(file?.bytes.length ?? 0));
    },
    stat(name: string, callback: Callback) {
      const bytes = lookup(name);
      done(callback, bytes ? null : error("ENOENT"), bytes ? stat(bytes.length) : undefined);
    },
    lstat(name: string, callback: Callback) { filesystem.stat(name, callback); },
    writeSync(descriptor: number, bytes: Uint8Array) {
      if (descriptor !== 1 && descriptor !== 2) throw error("EROFS");
      written += bytes.length;
      if (written > 65536) throw error("EFBIG");
      output(decoder.decode(bytes, { stream: true }));
      return bytes.length;
    },
    write(descriptor: number, bytes: Uint8Array, offset: number, length: number, _position: number | null, callback: Callback) {
      try { done(callback, null, filesystem.writeSync(descriptor, bytes.subarray(offset, offset + length))); }
      catch (failure) { done(callback, failure as Error); }
    },
    fsync(_descriptor: number, callback: Callback) { done(callback, null); },
  };
  return filesystem;
}
