import { describe, it, expect } from "vitest";
import { cn } from "./utils";

describe("utils", () => {
  describe("cn", () => {
    it("should merge class names correctly", () => {
      const result = cn("foo", "bar");
      expect(result).toBe("foo bar");
    });

    it("should handle conditional classes", () => {
      const result = cn("foo", false && "bar", "baz");
      expect(result).toBe("foo baz");
    });

    it("should merge tailwind classes correctly", () => {
      const result = cn("px-2 py-1", "px-4");
      expect(result).toBe("py-1 px-4");
    });

    it("should handle arrays of classes", () => {
      const result = cn(["foo", "bar"], "baz");
      expect(result).toBe("foo bar baz");
    });

    it("should handle empty input", () => {
      const result = cn();
      expect(result).toBe("");
    });

    it("should handle undefined and null", () => {
      const result = cn(undefined, null, "foo");
      expect(result).toBe("foo");
    });

    it("should merge complex tailwind utilities", () => {
      const result = cn("bg-red-500 hover:bg-blue-500", "bg-green-500");
      expect(result).toBe("hover:bg-blue-500 bg-green-500");
    });
  });
});
