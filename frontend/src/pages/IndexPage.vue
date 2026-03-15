<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { useOgpHead } from '@/composables/useOgpHead';
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import communicationBg from '@/assets/communication.png';
import phoneBg from '@/assets/phone.png';
import logoImage from '@/assets/logo.png';

const router = useRouter();
const playerName = ref('');
const loading = ref(false);
const errorMessage = ref('');

const isMobile = ref(false);

let mobileQuery: MediaQueryList | null = null;

function onQueryChange(e: MediaQueryListEvent) {
  isMobile.value = e.matches;
}

onMounted(() => {
  mobileQuery = window.matchMedia('(max-width: 640px)');
  isMobile.value = mobileQuery.matches;
  mobileQuery.addEventListener('change', onQueryChange);
});

onUnmounted(() => {
  mobileQuery?.removeEventListener('change', onQueryChange);
});

const bgImage = computed(() =>
  isMobile.value ? phoneBg : communicationBg,
);

useOgpHead(
  'InterKnot | トップページ',
  'InterKnotのトップページです。名前を入力してマッチングを始められます。',
);

const showRuleModal = ref(false);

function openRuleModal() {
  showRuleModal.value = true;
}

function closeRuleModal() {
  showRuleModal.value = false;
}

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
            :disabled="loading"
            class="rounded-xl border border-white/30 bg-white/80 px-4 py-3 text-center text-gray-900 placeholder-gray-400 outline-none backdrop-blur-sm focus:border-white/60 focus:bg-white/90 disabled:opacity-60"
          >
          <p
            v-if="errorMessage"
            class="text-sm text-red-300"
          >
            {{ errorMessage }}
          </p>

          <div class="flex w-full items-center gap-2">
            <button
              :disabled="playerName.trim() === '' || loading"
              class="flex-1 rounded-xl bg-emerald-500 py-3 font-semibold text-white shadow transition-colors hover:bg-emerald-400 active:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-40"
              @click="startMatching"
            >
              {{ loading ? 'マッチング中...' : 'マッチングをする' }}
            </button>
            <button
              class="rounded-xl bg-white/90 px-4 py-3 text-lg font-bold shadow transition-colors hover:bg-slate-100 active:bg-slate-200"
              aria-label="ルール説明"
              @click="openRuleModal"
            >
              ？
            </button>
          </div>

          <div
            v-if="showRuleModal"
            class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
            @click="closeRuleModal"
          >
            <div
              class="relative w-full max-w-md rounded-xl bg-white p-6 shadow-lg"
              @click.stop
            >
              <h2 class="mb-4 text-xl font-bold">
                ルール説明
              </h2>
              <ul class="mb-4 list-disc pl-5 text-left">
                <li>プレイヤー同士でマッチングしてゲームを開始します。</li>
                <li>ゲームの進行や操作方法は画面の指示に従ってください。</li>
                <li>不明点があればサポートまでお問い合わせください。</li>
              </ul>
              <button
                class="absolute right-2 top-2 text-slate-500 hover:text-slate-700"
                aria-label="閉じる"
                @click="closeRuleModal"
              >
                ×
              </button>
            </div>
          </div>
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
