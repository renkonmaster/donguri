<script setup lang="ts">
import { ref, computed } from 'vue';
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

const currentMessages = computed(() => {
  if (!selectedPlayerId.value) return [];
  return props.messages.filter(
    m =>
      (m.senderId === props.myPlayerId && m.receiverId === selectedPlayerId.value)
      || (m.senderId === selectedPlayerId.value && m.receiverId === props.myPlayerId),
  );
});

function selectPlayer(playerId: string) {
  selectedPlayerId.value = playerId;
  inputText.value = '';
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
  <div class="flex h-full overflow-hidden rounded-xl border border-gray-200 bg-white">
    <aside class="flex w-48 flex-col border-r border-gray-200 bg-gray-50">
      <div class="border-b border-gray-200 px-3 py-2 text-xs font-semibold text-gray-500">
        プレイヤー
      </div>
      <ul class="flex-1 overflow-y-auto">
        <li
          v-for="player in players"
          :key="player.id"
          class="flex cursor-pointer flex-col gap-0.5 px-3 py-2 transition-colors"
          :class="{
            'bg-blue-50': selectedPlayerId === player.id,
            'hover:bg-gray-100': selectedPlayerId !== player.id,
            'opacity-40': player.id === myPlayerId,
          }"
          @click="player.id !== myPlayerId && selectPlayer(player.id)"
        >
          <span class="flex items-center gap-1 truncate text-sm font-medium text-gray-800">
            <span
              class="inline-block h-2 w-2 rounded-full"
              :class="isAdjacent(player) ? 'bg-green-400' : 'bg-gray-300'"
              :title="isAdjacent(player) ? '接続中' : ''"
            />
            {{ player.name }}
            <span
              v-if="player.id === myPlayerId"
              class="text-xs text-gray-400"
            >(自分)</span>
          </span>
        </li>
      </ul>
    </aside>

    <div class="flex flex-1 flex-col">
      <div class="flex items-center justify-between border-b border-gray-200 px-4 py-2">
        <span class="font-semibold text-gray-700">
          {{ selectedPlayer ? selectedPlayer.name : '相手を選択してください' }}
        </span>
        <button
          v-if="selectedPlayer && isAdjacent(selectedPlayer)"
          class="rounded px-3 py-1 text-sm font-medium transition-colors"
          :class="
            swapRequested.has(selectedPlayer.id)
              ? 'bg-orange-100 text-orange-700 hover:bg-orange-200'
              : 'bg-blue-100 text-blue-700 hover:bg-blue-200'
          "
          @click="toggleSwap(selectedPlayer.id)"
        >
          {{ swapRequested.has(selectedPlayer.id) ? '交換申請中…' : '交換を申請' }}
        </button>
      </div>

      <div class="flex-1 overflow-y-auto px-4 py-3">
        <template v-if="selectedPlayerId">
          <p
            v-if="currentMessages.length === 0"
            class="text-center text-sm text-gray-400"
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
                class="max-w-[70%] rounded-2xl px-3 py-1.5 text-sm"
                :class="
                  msg.senderId === myPlayerId
                    ? 'rounded-br-sm bg-blue-500 text-white'
                    : 'rounded-bl-sm bg-gray-200 text-gray-800'
                "
              >
                <p>{{ msg.content }}</p>
                <p
                  class="mt-0.5 text-right text-[10px]"
                  :class="msg.senderId === myPlayerId ? 'text-blue-200' : 'text-gray-400'"
                >
                  {{ formatTime(msg.createdAt) }}
                </p>
              </div>
            </div>
          </div>
        </template>
        <p
          v-else
          class="text-center text-sm text-gray-400"
        >
          左のリストから相手を選んでください
        </p>
      </div>

      <div class="flex items-center gap-2 border-t border-gray-200 px-3 py-2">
        <input
          v-model="inputText"
          type="text"
          class="flex-1 rounded-lg border border-gray-300 px-3 py-1.5 text-sm outline-none focus:border-blue-400 disabled:bg-gray-100 disabled:text-gray-400"
          :placeholder="canChat ? 'メッセージを入力…' : '隣接していない相手とは話せません'"
          :disabled="!canChat"
          @keydown.enter="sendMessage"
        >
        <button
          class="rounded-lg bg-blue-500 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-blue-600 disabled:bg-gray-200 disabled:text-gray-400"
          :disabled="!canChat || !inputText.trim()"
          @click="sendMessage"
        >
          送信
        </button>
      </div>
    </div>
  </div>
</template>
