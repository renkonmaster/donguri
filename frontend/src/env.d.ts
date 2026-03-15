// src/env.d.ts
/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_PUBLIC_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}