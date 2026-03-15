<script setup lang="ts">
import { ref } from 'vue';
import DefaultLayout from '@/layouts/DefaultLayout.vue';

const playerName = ref('');
</script>

<template>
  <DefaultLayout>
    <div class="page-bg relative min-h-screen">
      <div class="absolute inset-0 bg-black/40" />
      <div class="relative flex min-h-screen flex-col items-center justify-center gap-6 px-6 pb-56 text-center">
        <h1 class="text-5xl font-bold tracking-widest text-white drop-shadow-lg">
          InterKnot
        </h1>
        <p class="text-white/70">絡まった糸をほどく、位置情報ゲーム</p>

        <div class="mx-auto mt-2 flex w-full max-w-xs flex-col gap-4">
          <input
            v-model="playerName"
            type="text"
            placeholder="あなたの名前"
            maxlength="20"
            class="rounded-xl border border-white/30 bg-white/20 px-4 py-3 text-center text-white placeholder-white/50 outline-none backdrop-blur-sm focus:border-white/60 focus:bg-white/25"
          />
          <button
            :disabled="playerName.trim() === ''"
            class="rounded-xl bg-emerald-500 py-3 font-semibold text-white shadow transition-colors hover:bg-emerald-400 active:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-40"
          >
            マッチングを始める
          </button>
        </div>
      </div>
    </div>
  </DefaultLayout>
</template>

<style scoped>
.page-bg {
  background-image: url('/communication.png');
  background-size: cover;
  background-position: calc(50% - 2px) center;
}

@media (max-width: 640px) {
  .page-bg {
    background-image: url('/正方形.png');
    background-position: center;
    overflow-x: hidden;
  }
}
</style>
