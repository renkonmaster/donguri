<script setup lang="ts">
import DefaultLayout from '@/layouts/DefaultLayout.vue';
</script>

<template>
  <DefaultLayout>
    <div class="flex flex-col items-center justify-center gap-8 py-24">
      <h1 class="text-6xl font-bold tracking-tight">どんぐり</h1>
      <button
        class="rounded-full bg-emerald-600 px-10 py-4 text-lg font-semibold text-white transition-colors hover:bg-emerald-500 active:bg-emerald-700"
      >
        マッチングを始める
      </button>
    </div>
  </DefaultLayout>
</template>
