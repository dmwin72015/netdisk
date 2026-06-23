/// <reference types="vitest/config" />
import { execSync } from "node:child_process";
import { sveltekit } from "@sveltejs/kit/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import { paraglideVitePlugin } from "@inlang/paraglide-js";

function getGitVersion() {
  if (process.env.APP_VERSION) return process.env.APP_VERSION;
  try {
    const tag = execSync("git describe --tags --abbrev=0 2>/dev/null || true").toString().trim();
    const sha = execSync("git rev-parse --short HEAD 2>/dev/null || true").toString().trim();
    if (tag && sha) return `${tag}-${sha}`;
    if (tag) return tag;
    if (sha) return sha;
  } catch {}
  return "";
}

const buildTime = new Date().toISOString();
const buildStamp = buildTime.replace(/[-:TZ.]/g, "").slice(0, 14);
const gitVersion = getGitVersion();
const appVersion = gitVersion ? `${buildStamp}-${gitVersion}` : buildStamp;

export default defineConfig(({ mode }) => {
  const isProduction = mode === "production";

  return {
    define: {
      __APP_VERSION__: JSON.stringify(appVersion),
      __APP_BUILD_TIME__: JSON.stringify(buildTime),
    },
    plugins: [
      tailwindcss(),
      sveltekit(),
      paraglideVitePlugin({
        project: "./project.inlang",
        outdir: "./src/lib/paraglide",
        strategy: ["cookie", "baseLocale"],
      }),
    ],
    ssr: {
      noExternal: ['@lucide/svelte'],
      optimizeDeps: {
        include: ['@lucide/svelte'],
      },
    },
    build: {
      rolldownOptions: {
        output: {
          minify: isProduction
            ? {
                compress: {
                  dropConsole: true,
                  dropDebugger: true,
                },
              }
            : "dce-only",
        },
        checks: {
          pluginTimings: false,
        },
      },
    },

    server: {
      host: "0.0.0.0",
      port: 5173,
      proxy: {
        "/api": { target: "http://localhost:8080", changeOrigin: true },
        "/hls": { target: "http://localhost:8080", changeOrigin: true },
      },
    },
    test: {
      include: ["tests/**/*.test.ts"],
      globals: true,
      setupFiles: ['./tests/setup.ts'],
    },
  };
});
