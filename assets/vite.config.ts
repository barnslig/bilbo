import { defineConfig } from "vite";

export default defineConfig({
  build: {
    manifest: true,
    rollupOptions: {
      input: ["src/app.js", "src/editor.js"],
    },
  },
});
