<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
import GameMap from '@/components/GameMap.vue';
import ChatOverlay from '@/components/ChatOverlay.vue';
import type { MapClickPayload, MapPoint } from '@/types/map';
import type { Player, Message } from '@/types/game';
import { idToRgb } from '@/utils/pointColor';

const props = defineProps<{
  myPlayerId: string;
  players: Player[];
  messages: Message[];
  roomStatus: 'matching' | 'playing' | 'finished';
  swapCount: number;
  clearTimeMs: number | null;
}>();

const emit = defineEmits<{
  sendMessage: [receiverId: string, content: string];
  toggleSwap: [targetPlayerId: string, needsSwap: boolean];
}>();

const selectedPlayerId = ref<string | null>(null);
const swapRequested = reactive(new Set<string>());

const me = computed(() => props.players.find(p => p.id === props.myPlayerId));

function isAdjacent(player: Player): boolean {
  if (!me.value) return false;
  const n = props.players.length;
  const diff = Math.abs(player.orderIndex - me.value.orderIndex);
  return diff === 1 || diff === n - 1;
}

const selectedPlayer = computed(() =>
  props.players.find(p => p.id === selectedPlayerId.value) ?? null,
);

const canChat = computed(() =>
  selectedPlayer.value != null && isAdjacent(selectedPlayer.value),
);

const adjacentPlayers = computed(() =>
  props.players.filter(p => p.id !== props.myPlayerId && isAdjacent(p)),
);

const currentMessages = computed(() => {
  if (!selectedPlayerId.value) return [];
  return props.messages.filter(
    m =>
      (m.senderId === props.myPlayerId && m.receiverId === selectedPlayerId.value)
      || (m.senderId === selectedPlayerId.value && m.receiverId === props.myPlayerId),
  );
});

// players の orderIndex が変化した (= スワップ成立) ときに swapRequested をクリアする
watch(
  () => props.players.map(p => ({ id: p.id, orderIndex: p.orderIndex })),
  (newVal, oldVal) => {
    if (!oldVal) return;
    for (const newP of newVal) {
      const oldP = oldVal.find(p => p.id === newP.id);
      if (oldP && oldP.orderIndex !== newP.orderIndex) {
        swapRequested.delete(newP.id);
      }
    }
  },
  { deep: true },
);

// playerId → その会話で既読済みの inbound (相手→自分) メッセージ件数
const seenCount = reactive(new Map<string, number>());

// オーバーレイが開いている間・新着が届いた瞬間も既読にする (inbound 件数ベース)
watch([selectedPlayerId, currentMessages], () => {
  const id = selectedPlayerId.value;
  if (id) {
    const inbound = currentMessages.value.filter(m => m.senderId === id).length;
    seenCount.set(id, inbound);
  }
});

const unreadPlayerIds = computed(() =>
  props.players
    .filter(p => p.id !== props.myPlayerId)
    .filter((p) => {
      const inbound = props.messages.filter(
        m => m.senderId === p.id && m.receiverId === props.myPlayerId,
      ).length;
      return inbound > (seenCount.get(p.id) ?? 0);
    })
    .map(p => p.id),
);

// GameMap は order_index 順に並んだ MapPoint[] を受け取る
const mapPoints = computed<MapPoint[]>(() =>
  [...props.players]
    .sort((a, b) => a.orderIndex - b.orderIndex)
    .map(p => ({ id: p.id, orderIndex: p.orderIndex, lat: p.lat, lng: p.lng, name: p.name })),
);

const formattedClearTime = computed(() => {
  if (props.clearTimeMs === null) return '--:--';
  const totalSeconds = Math.floor(props.clearTimeMs / 1000);
  const m = Math.floor(totalSeconds / 60).toString().padStart(2, '0');
  const s = (totalSeconds % 60).toString().padStart(2, '0');
  return `${m}:${s}`;
});

const twitterShareUrl = computed(() => {
  const text = `InterKnot で ${props.players.length} 人と一緒に糸をほどきました！ クリアタイム: ${formattedClearTime.value}（交換 ${props.swapCount} 回）`;
  return `https://twitter.com/intent/tweet?text=${encodeURIComponent(text)}&hashtags=InterKnot`;
});

function playerBadgeStyle(player: Player): Record<string, string> {
  const [r, g, b] = idToRgb(String(player.orderIndex), 'highlight');
  return { backgroundColor: `rgb(${r} ${g} ${b})` };
}

function onMapClick(payload: MapClickPayload) {
  if (!payload.point || payload.point.id === props.myPlayerId) return;
  selectedPlayerId.value = payload.point.id;
}

function onSendMessage(content: string) {
  if (!selectedPlayerId.value) return;
  emit('sendMessage', selectedPlayerId.value, content);
}

function onToggleSwap() {
  if (!selectedPlayerId.value) return;
  const id = selectedPlayerId.value;
  const next = !swapRequested.has(id);
  if (next) {
    swapRequested.add(id);
  }
  else {
    swapRequested.delete(id);
  }
  emit('toggleSwap', id, next);
}
</script>

<template>
  <div class="relative h-full">
    <GameMap
      :points="mapPoints"
      :highlighted-id="myPlayerId"
      :unread-ids="unreadPlayerIds"
      show-line
      @click="onMapClick"
    />

    <!-- ヒント（チャット未選択時） -->
    <Transition name="fade">
      <div
        v-if="!selectedPlayer"
        class="pointer-events-none absolute inset-x-0 bottom-0 z-10 flex justify-center px-4 pb-6 pt-3"
      >
        <div class="w-full max-w-lg rounded-2xl bg-black/20 px-4 py-3 text-center text-sm text-gray-900/80 backdrop-blur-[2px]">
          <p>
            隣り合っている人
            <template v-if="adjacentPlayers.length > 0">
              ({{ adjacentPlayers.map(p => p.name).join(' さんと ') }} さん)
            </template>
            とのみ会話が可能です
          </p>
          <p class="mt-1">
            人の点を押すことで、その人とのチャット・交換ボタン画面を表示することができます
          </p>
        </div>
      </div>
    </Transition>

    <!-- チャットオーバーレイ -->
    <Transition name="fade">
      <ChatOverlay
        v-if="selectedPlayer"
        :my-player-id="myPlayerId"
        :player="selectedPlayer"
        :can-chat="canChat"
        :messages="currentMessages"
        :swap-requested="swapRequested.has(selectedPlayer.id)"
        @close="selectedPlayerId = null"
        @send-message="onSendMessage"
        @toggle-swap="onToggleSwap"
      />
    </Transition>

    <!-- ゲームクリアオーバーレイ -->
    <Transition name="clear">
      <div
        v-if="roomStatus === 'finished'"
        class="absolute inset-0 z-30 flex items-end justify-center bg-black/50 backdrop-blur-[2px] pb-8 px-4"
      >
        <div class="clear-card w-full max-w-sm rounded-3xl bg-white shadow-2xl overflow-hidden">
          <!-- ヘッダー -->
          <div class="bg-gradient-to-br from-emerald-400 to-teal-500 px-6 pt-7 pb-5 text-center text-white">
            <p class="text-4xl mb-1">
              &#127881;
            </p>
            <h1 class="text-3xl font-black tracking-tight">
              ほどけた！
            </h1>
            <p class="mt-1 text-emerald-100 text-sm font-medium">
              {{ players.length }} 人で糸を解きほぐしました
            </p>
          </div>

          <!-- スタッツ -->
          <div class="grid grid-cols-3 divide-x divide-gray-100 border-b border-gray-100">
            <div class="py-4 text-center">
              <p class="text-xs text-gray-400 mb-0.5">
                タイム
              </p>
              <p class="text-xl font-bold text-gray-800 tabular-nums">
                {{ formattedClearTime }}
              </p>
            </div>
            <div class="py-4 text-center">
              <p class="text-xs text-gray-400 mb-0.5">
                交換回数
              </p>
              <p class="text-xl font-bold text-gray-800">
                {{ swapCount }}<span class="text-sm font-normal text-gray-500">回</span>
              </p>
            </div>
            <div class="py-4 text-center">
              <p class="text-xs text-gray-400 mb-0.5">
                参加人数
              </p>
              <p class="text-xl font-bold text-gray-800">
                {{ players.length }}<span class="text-sm font-normal text-gray-500">人</span>
              </p>
            </div>
          </div>

          <!-- プレイヤーリスト -->
          <div class="px-5 pt-4 pb-3">
            <p class="text-xs font-semibold tracking-widest text-gray-400 mb-2 uppercase">
              参加者
            </p>
            <div class="flex flex-wrap gap-1.5">
              <span
                v-for="p in [...players].sort((a, b) => a.orderIndex - b.orderIndex)"
                :key="p.id"
                class="rounded-full px-3 py-1 text-sm font-semibold text-white"
                :style="playerBadgeStyle(p)"
              >
                {{ p.name }}{{ p.id === myPlayerId ? ' ★' : '' }}
              </span>
            </div>
          </div>

          <!-- ボタン -->
          <div class="px-5 pb-6 pt-2 flex flex-col gap-2">
            <a
              :href="twitterShareUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="flex items-center justify-center gap-2 rounded-xl bg-black py-3 text-sm font-bold text-white transition-opacity hover:opacity-75"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                fill="currentColor"
                class="h-4 w-4 shrink-0"
              >
                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-4.714-6.231-5.401 6.231H2.746l7.73-8.835L1.254 2.25H8.08l4.26 5.625L18.245 2.25zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
              </svg>
              X でシェア
            </a>
            <a
              href="/"
              class="flex items-center justify-center rounded-xl border-2 border-gray-200 py-3 text-sm font-bold text-gray-600 transition-colors hover:bg-gray-50"
            >
              トップページへ戻る
            </a>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.clear-enter-active {
  transition: opacity 0.4s ease;
}

.clear-enter-active .clear-card {
  transition: opacity 0.4s ease 0.1s, transform 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) 0.1s;
}

.clear-enter-from {
  opacity: 0;
}

.clear-enter-from .clear-card {
  opacity: 0;
  transform: translateY(40px) scale(0.95);
}
</style>
