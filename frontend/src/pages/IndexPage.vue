<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useHead } from '@unhead/vue';
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import communicationBg from '@/assets/communication.png';
import phoneBg from '@/assets/phone.png';
import logoImage from '@/assets/logo.png';

const router = useRouter();
const route = useRoute();
const playerName = ref('');

const baseUrl = import.meta.env.VITE_PUBLIC_URL;
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

useHead({
  title: 'InterKnot | トップページ',
  meta: [
    { name: 'description', content: 'InterKnotのトップページです。名前を入力してマッチングを始められます。' },

    { property: 'og:title', content: 'InterKnot | トップページ' },
    { property: 'og:description', content: 'InterKnotのトップページです。名前を入力してマッチングを始められます。' },
    { property: 'og:type', content: 'website' },
    { property: 'og:url', content: `${baseUrl}${route.path}` },
    { property: 'og:image', content: `${baseUrl}/ogp.png` },

    { name: 'twitter:card', content: 'summary_large_image' },
    { name: 'twitter:title', content: 'InterKnot | トップページ' },
    { name: 'twitter:description', content: 'InterKnotのトップページです。名前を入力してマッチングを始められます。' },
    { name: 'twitter:image', content: `${baseUrl}/ogp.png` },
  ],
  link: [
    { rel: 'canonical', href: `${baseUrl}${route.path}` },
  ],
});

const showRuleModal = ref(false);

function openRuleModal() {
  showRuleModal.value = true;
}

function closeRuleModal() {
  showRuleModal.value = false;
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
            class="rounded-xl border border-white/70 bg-white/90 px-4 py-3 text-center text-slate-700 placeholder-slate-400 outline-none shadow-md backdrop-blur-sm focus:border-sky-300 focus:bg-white"
          >

          <div class="flex w-full items-center gap-2">
            <button
              :disabled="playerName.trim() === ''"
              class="flex-1 rounded-xl bg-emerald-500 py-3 font-semibold text-white shadow transition-colors hover:bg-emerald-400 active:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-40"
              @click="router.push('/game')"
            >
              マッチングをする
            </button>
            <button
              class="rounded-xl bg-white/90 py-3 px-4 flex items-center justify-center text-lg font-bold shadow hover:bg-slate-100 active:bg-slate-200"
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
              class="bg-white rounded-xl shadow-lg p-6 max-w-md w-full relative"
              @click.stop
            >
              <h2 class="text-xl font-bold mb-4">
                ルール説明
              </h2>
              <ul class="text-left list-disc pl-5 mb-4">
                <li>プレイヤー同士でマッチングしてゲームを開始します。</li>
                <li>ゲームの進行や操作方法は画面の指示に従ってください。</li>
                <li>不明点があればサポートまでお問い合わせください。</li>
              </ul>
              <button
                class="absolute top-2 right-2 text-slate-500 hover:text-slate-700"
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
