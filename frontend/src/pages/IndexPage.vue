<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import communicationBg from '@/assets/communication.png';
import phoneBg from '@/assets/phone.png';
import logoImage from '@/assets/logo.png';

const router = useRouter();
const playerName = ref('');

const mobileQuery = window.matchMedia('(max-width: 640px)');
const isMobile = ref(mobileQuery.matches);

function onQueryChange(e: MediaQueryListEvent) {
  isMobile.value = e.matches;
}

onMounted(() => mobileQuery.addEventListener('change', onQueryChange));
onUnmounted(() => mobileQuery.removeEventListener('change', onQueryChange));

const bgImage = computed(() =>
  isMobile.value ? phoneBg : communicationBg,
);
</script>

<template>
  <DefaultLayout>
    <div
      class="page-bg relative min-h-screen"
      :style="{ backgroundImage: `url(${bgImage})` }"
    >
      <div class="relative flex min-h-screen flex-col items-center justify-center px-6 pb-56 text-center">
        <div class="flex flex-col items-center gap-4">
          <img
            :src="logoImage"
            alt="InterKnot logo"
            class="w-56 sm:w-72 md:w-80 h-auto object-contain drop-shadow-lg"
          >
        </div>

        <div class="mx-auto mt-8 flex w-full max-w-xs flex-col gap-4">
          <label
            for="player-name"
            class="sr-only"
          >
            名前
          </label>

          <input
            id="player-name"
            v-model="playerName"
            type="text"
            placeholder="あなたの名前"
            maxlength="20"
            class="rounded-xl border border-white/70 bg-white/90 px-4 py-3 text-center text-slate-700 placeholder-slate-400 outline-none shadow-md backdrop-blur-sm focus:border-sky-300 focus:bg-white"
          >

          <button
            :disabled="playerName.trim() === ''"
            class="rounded-xl bg-emerald-500 py-3 font-semibold text-white shadow transition-colors hover:bg-emerald-400 active:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-40"
            @click="router.push('/game')"
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
  background-size: cover;
  background-position: calc(50% - 2px) center;
}

@media (width <= 640px) {
  .page-bg {
    background-position: center;
    overflow-x: hidden;
  }
}
</style>