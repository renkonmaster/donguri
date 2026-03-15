<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { useOgpHead } from '@/composables/useOgpHead';
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import communicationBg from '@/assets/communication.png';
import phoneBg from '@/assets/phone.png';
import logoImage from '@/assets/logo.png';
import slide13Image from '@/assets/modal-images/Slide13.jpg';
import slide14Image from '@/assets/modal-images/Slide14.jpg';
import slide15Image from '@/assets/modal-images/Slide15.jpg';
import slide16Image from '@/assets/modal-images/Slide16.jpg';

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
const currentRuleSlideIndex = ref(0);

const ruleSlides = [
  {
    image: slide13Image,
    paragraphs: [
      'あなたの名前にハンドルネームを入力してください。',
      'ほかのユーザーの画面にその名前が表示されます。',
      'マッチングボタンを押して、8人集まるとゲームが始まります。',
      '「？」を押すとゲーム説明がでてきます。',
    ],
  },
  {
    image: slide14Image,
    paragraphs: [
      'ゲームが始まると初期のループが描かれています。',
      'このループに交差ができないように点をほかの人と交換していきましょう。',
    ],
  },
  {
    image: slide15Image,
    paragraphs: [
      'ほかの人の現在位置をタップするとメッセージ画面がでてきます。',
      'メッセージができるのは自分と線がつながっている人のみです。',
      '線がつながっている人からメッセージが届いた場合は右上図のようにその人に目印がつきます。',
    ],
  },
  {
    image: slide16Image,
    paragraphs: [
      '線がつながっている人同士は線のつながりの順番を交換することができます。上部の「交換する」ボタンを押すと「交換申請中」に代わります。相手もボタンを押し「交換申請中」になったときに交換が成立します。',
      '相手に自分が交換申請をしていることは伝わらないので、チャットを通じて交換を成立させましょう。',
    ],
  },
];

const currentRuleSlide = computed(() => ruleSlides[currentRuleSlideIndex.value]);
const isFirstRuleSlide = computed(() => currentRuleSlideIndex.value === 0);
const isLastRuleSlide = computed(() => currentRuleSlideIndex.value === ruleSlides.length - 1);

function openRuleModal() {
  currentRuleSlideIndex.value = 0;
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
  }
  catch (e) {
    errorMessage.value = e instanceof Error ? e.message : 'エラーが発生しました';
    loading.value = false;
  }
}

function goToNextRuleSlide() {
  if (isLastRuleSlide.value) {
    closeRuleModal();
    return;
  }
  currentRuleSlideIndex.value += 1;
}

function goToPrevRuleSlide() {
  if (isFirstRuleSlide.value) {
    return;
  }
  currentRuleSlideIndex.value -= 1;
}

function goToRuleSlide(index: number) {
  currentRuleSlideIndex.value = index;
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
            class="fixed inset-0 z-50 flex items-center justify-center bg-black/55 px-4"
            @click="closeRuleModal"
          >
            <div
              class="relative w-full max-w-lg rounded-2xl bg-white p-5 shadow-xl sm:p-6"
              @click.stop
            >
              <div class="mb-4 flex items-start justify-between gap-3">
                <h2 class="text-lg font-bold text-slate-800 sm:text-xl">
                  ゲーム説明
                </h2>
                <p class="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold text-slate-600 sm:text-sm">
                  {{ currentRuleSlideIndex + 1 }} / {{ ruleSlides.length }}
                </p>
              </div>

              <img
                :src="currentRuleSlide.image"
                :alt="`説明スライド ${currentRuleSlideIndex + 1}`"
                class="mb-4 h-48 w-full rounded-xl object-cover sm:h-56"
              >

              <div class="space-y-3 text-left text-sm leading-relaxed text-slate-700 sm:text-base">
                <p
                  v-for="(paragraph, paragraphIndex) in currentRuleSlide.paragraphs"
                  :key="`${currentRuleSlideIndex}-${paragraphIndex}`"
                >
                  {{ paragraph }}
                </p>
              </div>

              <div class="mt-5 flex items-center justify-center gap-2">
                <button
                  v-for="(_, index) in ruleSlides"
                  :key="`rule-dot-${index}`"
                  class="h-2.5 w-2.5 rounded-full transition-colors"
                  :class="index === currentRuleSlideIndex ? 'bg-emerald-500' : 'bg-slate-300 hover:bg-slate-400'"
                  :aria-label="`スライド ${index + 1} を表示`"
                  @click="goToRuleSlide(index)"
                />
              </div>

              <div class="mt-6 flex items-center justify-between gap-3">
                <button
                  class="rounded-lg border border-slate-300 px-4 py-2 text-sm font-semibold text-slate-700 transition hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-40"
                  :disabled="isFirstRuleSlide"
                  @click="goToPrevRuleSlide"
                >
                  前へ
                </button>
                <button
                  class="rounded-lg bg-emerald-500 px-4 py-2 text-sm font-semibold text-white transition hover:bg-emerald-400"
                  @click="goToNextRuleSlide"
                >
                  {{ isLastRuleSlide ? '閉じる' : '次へ' }}
                </button>
              </div>

              <button
                class="absolute right-3 top-3 text-slate-500 hover:text-slate-700"
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
