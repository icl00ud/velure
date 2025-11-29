import path from "node:path";
import react from "@vitejs/plugin-react-swc";
import { componentTagger } from "lovable-tagger";
import { defineConfig } from "vite";

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => ({
  server: {
    host: "::",
    port: 8080,
  },
  plugins: [react(), mode === "development" && componentTagger()].filter(Boolean),
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  test: {
    globals: true,
    environment: "happy-dom",
    setupFiles: ["./src/test/setup.ts"],
    css: true,
    coverage: {
      provider: "istanbul",
      reporter: ["text", "json", "html", "lcov"],
      all: true,
      include: ["src/**/*.{ts,tsx}"],
      exclude: [
        "node_modules/",
        "src/components/ui/**",
        "src/test/",
        "**/*.d.ts",
        "**/*.config.*",
        "**/mockData",
        "**/*.test.{ts,tsx}",
        "**/*.spec.{ts,tsx}",
      ],
    },
  },
}));
