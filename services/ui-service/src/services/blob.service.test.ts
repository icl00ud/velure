import { blobService } from "./blob.service";
import { vi } from "vitest";

const mockFetch = (response: Partial<Response>) => {
  global.fetch = vi.fn().mockResolvedValue({
    ok: true,
    blob: vi.fn().mockResolvedValue(new Blob(["data"])),
    ...response,
  } as any);
};

describe("BlobService", () => {
  const originalFileReader = global.FileReader;
  const originalCreateObjectURL = URL.createObjectURL;
  const originalRevokeObjectURL = URL.revokeObjectURL;
  const originalCreateElement = document.createElement;

  beforeEach(() => {
    vi.restoreAllMocks();
  });

  afterEach(() => {
    global.FileReader = originalFileReader;
    URL.createObjectURL = originalCreateObjectURL;
    URL.revokeObjectURL = originalRevokeObjectURL;
    document.createElement = originalCreateElement;
  });

  it("converts a blob URL to base64", async () => {
    mockFetch({});

    const base64String = "data:text/plain;base64,ZGF0YQ==";

    // Stub FileReader to control the read result
    global.FileReader = class {
      result: string | ArrayBuffer | null = base64String;
      onload: (() => void) | null = null;
      onerror: ((error: unknown) => void) | null = null;
      readAsDataURL() {
        this.onload?.();
      }
    } as any;

    const result = await blobService.getBase64FromUrl("https://example.com/file.txt");

    expect(result).toBe("ZGF0YQ==");
    expect(fetch).toHaveBeenCalledWith("https://example.com/file.txt");
  });

  it("throws when blob download fails", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      blob: vi.fn(),
    } as any);

    await expect(blobService.getBase64FromUrl("https://example.com/file.txt")).rejects.toThrow();
  });

  it("downloads a blob and triggers browser download", async () => {
    const anchor = document.createElement("a");
    anchor.click = vi.fn();

    const createUrlSpy = vi.spyOn(URL, "createObjectURL").mockReturnValue("blob:url");
    const revokeSpy = vi.spyOn(URL, "revokeObjectURL");

    const originalCreator = document.createElement.bind(document);
    vi.spyOn(document, "createElement").mockImplementation((tagName: string) => {
      if (tagName === "a") {
        return anchor;
      }
      return originalCreator(tagName);
    });

    mockFetch({});

    await blobService.downloadBlob("https://example.com/file.txt", "file.txt");

    expect(fetch).toHaveBeenCalledWith("https://example.com/file.txt");
    expect(createUrlSpy).toHaveBeenCalled();
    expect(anchor.href).toBe("blob:url");
    expect(anchor.download).toBe("file.txt");
    expect(anchor.click).toHaveBeenCalled();
    expect(revokeSpy).toHaveBeenCalledWith("blob:url");
  });

  it("throws when download response is not ok", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      blob: vi.fn(),
    } as any);

    await expect(blobService.downloadBlob("https://example.com/file.txt")).rejects.toThrow();
  });
});
