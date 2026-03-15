<script setup lang="ts">
import { ref, computed } from 'vue';
import GameMap from '@/components/GameMap.vue';
import type { MapClickPayload, MapPoint } from '@/types/map';
import type { Player, Message } from '@/types/game';

const props = defineProps<{
  myPlayerId: string;
  players: Player[];
  messages: Message[];
}>();

const emit = defineEmits<{
  sendMessage: [receiverId: string, content: string];
  toggleSwap: [targetPlayerId: string, needsSwap: boolean];
}>();

const selectedPlayerId = ref<string | null>(null);
const inputText = ref('');
const swapRequested = ref<Set<string>>(new Set());

const me = computed(() => props.players.find(p => p.id === props.myPlayerId));

function isAdjacent(player: Player): boolean {
  if (!me.value) return false;
  return Math.abs(player.orderIndex - me.value.orderIndex) === 1;
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

// GameMap は order_index 順に並んだ MapPoint[] を受け取る
const mapPoints = computed<MapPoint[]>(() =>
  [...props.players]
    .sort((a, b) => a.orderIndex - b.orderIndex)
    .map(p => ({ id: p.id, lat: p.lat, lng: p.lng, name: p.name })),
);

function onMapClick(payload: MapClickPayload) {
  if (!payload.point || payload.point.id === props.myPlayerId) return;
  selectedPlayerId.value = payload.point.id;
  inputText.value = '';
}

function closeChat() {
  selectedPlayerId.value = null;
}

function sendMessage() {
  const text = inputText.value.trim();
  if (!text || !selectedPlayerId.value || !canChat.value) return;
  emit('sendMessage', selectedPlayerId.value, text);
  inputText.value = '';
}

function toggleSwap(targetId: string) {
  const next = !swapRequested.value.has(targetId);
  if (next) {
    swapRequested.value.add(targetId);
  }
  else {
    swapRequested.value.delete(targetId);
  }
  // Set は参照が変わらないため再代入で reactivity を確保
  swapRequested.value = new Set(swapRequested.value);
  emit('toggleSwap', targetId, next);
}

function formatTime(date: Date): string {
  return date.toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' });
}
</script>

<template>
  <div class="relative h-full">
    <GameMap
      :points="mapPoints"
      :highlighted-id="myPlayerId"
      show-line
      @click="onMapClick"
    />

    <!-- ヒント（チャット未選択時） -->
    <Transition name="chat">
      <div
        v-if="!selectedPlayer"
        class="absolute inset-x-0 bottom-0 z-10 flex justify-center px-4 pb-6 pt-3 pointer-events-none"
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
    <Transition name="chat">
      <div
        v-if="selectedPlayer"
        class="absolute inset-0 z-10 flex justify-center bg-black/40"
      >
        <div class="flex w-full max-w-lg flex-col">
          <!-- ヘッダ -->
          <div class="flex shrink-0 items-center gap-3 px-4 py-3">
            <button
              class="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-gray-700/80 text-xl text-white backdrop-blur-sm transition-colors hover:bg-gray-600/80"
              @click="closeChat"
            >
              ✕
            </button>
            <button
              class="flex h-14 flex-1 items-center justify-center rounded-full px-4 text-base font-bold text-white backdrop-blur-sm transition-colors break-keep leading-tight"
              :class="
                !canChat
                  ? 'cursor-not-allowed bg-gray-500/60'
                  : swapRequested.has(selectedPlayer.id)
                    ? 'bg-orange-500/80 hover:bg-orange-400/80'
                    : 'bg-blue-500/80 hover:bg-blue-400/80'
              "
              :disabled="!canChat"
              @click="toggleSwap(selectedPlayer.id)"
            >
              <span v-if="!canChat">接続していない人と 交換することはできません</span>
              <span v-else>{{ swapRequested.has(selectedPlayer.id) ? '交換申請中…… (もう一度押すとキャンセル)' : '交換する' }}</span>
            </button>
          </div>

          <!-- メッセージ一覧 -->
          <div class="flex-1 overflow-y-auto px-4 py-2">
            <p
              v-if="currentMessages.length === 0"
              class="text-center text-base text-white/60"
            >
              まだメッセージはありません
            </p>
            <div
              v-else
              class="flex flex-col gap-2"
            >
              <div
                v-for="msg in currentMessages"
                :key="msg.id"
                class="flex"
                :class="msg.senderId === myPlayerId ? 'justify-end' : 'justify-start'"
              >
                <div
                  class="max-w-[70%] rounded-2xl px-4 py-2 text-base backdrop-blur-sm"
                  :class="
                    msg.senderId === myPlayerId
                      ? 'rounded-br-sm bg-blue-500/80 text-white'
                      : 'rounded-bl-sm bg-white/70 text-gray-900'
                  "
                >
                  <p>{{ msg.content }}</p>
                  <p
                    class="mt-0.5 text-right text-xs"
                    :class="msg.senderId === myPlayerId ? 'text-blue-200' : 'text-gray-500'"
                  >
                    {{ formatTime(msg.createdAt) }}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <!-- 入力欄 -->
          <div class="shrink-0 px-3 pb-5 pt-2">
            <div class="flex items-center gap-2 rounded-2xl bg-white/20 px-4 py-2.5 backdrop-blur-sm">
              <p
                v-if="!canChat"
                class="flex-1 text-base text-white/50"
              >
                {{ selectedPlayer.name }} さんとは隣り合っていません
              </p>
              <input
                v-else
                v-model="inputText"
                type="text"
                placeholder="メッセージを入力…"
                class="flex-1 bg-transparent text-base text-white placeholder-white/50 outline-none"
                @keydown.enter="sendMessage"
              >
              <button
                class="shrink-0 rounded-full bg-blue-500 px-4 py-2 text-base font-medium text-white transition-colors hover:bg-blue-400 disabled:opacity-30"
                :disabled="!canChat || !inputText.trim()"
                @click="sendMessage"
              >
                送信
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.chat-enter-active,
.chat-leave-active {
  transition: opacity 0.2s ease;
}

.chat-enter-from,
.chat-leave-to {
  opacity: 0;
}
</style>
