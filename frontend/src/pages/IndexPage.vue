<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import communicationBg from '@/assets/communication.png';
import phoneBg from '@/assets/phone.png';

const router = useRouter();
const playerName = ref('');
const loading = ref(false);
const errorMessage = ref('');

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

async function startMatching() {
  if (playerName.value.trim() === '') return;
  loading.value = true;
  errorMessage.value = '';
  try {
    // TODO: GPS 取得に切り替える
    const lat = Math.random() * 130 - 60;
    const lng = Math.random() * 360 - 180;
    const res = await fetch('/api/rooms/join', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: playerName.value.trim(), lat, lng }),
    });
    if (!res.ok) {
      const data = await res.json().catch(() => ({})) as { message?: string };
      throw new Error(data.message ?? 'サーバーエラーが発生しました');
    }
    const data = await res.json() as { room_id: string; player_id: string };
    await router.push({ path: '/game', query: { room_id: data.room_id, player_id: data.player_id } });
  } catch (e) {
    errorMessage.value = e instanceof Error ? e.message : 'エラーが発生しました';
    loading.value = false;
  }
}
</script>

<template>
  <DefaultLayout>
    <div
      class="page-bg relative min-h-screen"
      :style="{ backgroundImage: `url(${bgImage})` }"
    >
      <div class="absolute inset-0 bg-black/33" />
      <div class="relative flex min-h-screen flex-col items-center justify-center gap-6 px-6 pb-56 text-center">
        <h1 class="text-5xl font-bold tracking-widest text-white drop-shadow-lg">
          InterKnot
        </h1>
        <p class="text-white/70">
          絡まった糸をほどく、位置情報ゲーム
        </p>

        <div class="mx-auto mt-2 flex w-full max-w-xs flex-col gap-4">
          <label
            for="player-name"
            class="sr-only"
          >名前</label>
          <input
            id="player-name"
            v-model="playerName"
            type="text"
            placeholder="あなたの名前"
            maxlength="20"
            :disabled="loading"
            class="rounded-xl border border-white/30 bg-white/20 px-4 py-3 text-center text-white placeholder-white/50 outline-none backdrop-blur-sm focus:border-white/60 focus:bg-white/25 disabled:opacity-60"
          >
          <p
            v-if="errorMessage"
            class="text-sm text-red-300"
          >
            {{ errorMessage }}
          </p>
          <button
            :disabled="playerName.trim() === '' || loading"
            class="rounded-xl bg-emerald-500 py-3 font-semibold text-white shadow transition-colors hover:bg-emerald-400 active:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-40"
            @click="startMatching"
          >
            {{ loading ? 'マッチング中...' : 'マッチングを始める' }}
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
