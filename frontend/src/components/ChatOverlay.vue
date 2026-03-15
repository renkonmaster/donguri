<script setup lang="ts">
import { ref, computed } from 'vue';
import type { Player, Message } from '@/types/game';
import { idToRgb } from '@/utils/pointColor';

const props = defineProps<{
  myPlayerId: string;
  player: Player;
  canChat: boolean;
  messages: Message[];
  swapRequested: boolean;
}>();

const emit = defineEmits<{
  close: [];
  sendMessage: [content: string];
  toggleSwap: [];
}>();

const inputText = ref('');

// --pc-r/g/b を設定し、CSS 側で rgb(var(--pc-r) var(--pc-g) var(--pc-b) / alpha) として使用する
const themeVars = computed(() => {
  const [r, g, b] = idToRgb(props.player.id, 'highlight');
  return { '--pc-r': r, '--pc-g': g, '--pc-b': b };
});

function sendMessage() {
  const text = inputText.value.trim();
  if (!text || !props.canChat) return;
  emit('sendMessage', text);
  inputText.value = '';
}

function formatTime(date: Date): string {
  return date.toLocaleTimeString('ja-JP', { hour: '2-digit', minute: '2-digit' });
}
</script>

<template>
  <div
    class="absolute inset-0 z-10 flex justify-center bg-black/40"
    :style="themeVars"
  >
    <div class="flex w-full max-w-lg flex-col">
      <div class="flex shrink-0 items-center gap-3 px-4 py-3">
        <button
          class="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-gray-700/80 text-xl text-white backdrop-blur-sm transition-colors hover:bg-gray-600/80"
          @click="emit('close')"
        >
          ✕
        </button>
        <button
          class="flex h-14 flex-1 items-center justify-center rounded-full px-4 text-base font-bold text-white backdrop-blur-sm transition-colors break-keep leading-tight"
          :class="
            !canChat
              ? 'cursor-not-allowed bg-gray-500/60'
              : swapRequested
                ? 'bg-orange-500/80 hover:bg-orange-400/80'
                : 'theme-btn'
          "
          :disabled="!canChat"
          @click="emit('toggleSwap')"
        >
          <span v-if="!canChat">接続していない人と 交換することはできません</span>
          <span v-else>{{ swapRequested ? '交換申請中…… (もう一度押すとキャンセル)' : '交換する' }}</span>
        </button>
      </div>

      <div class="flex items-center justify-center gap-2 pb-1">
        <span class="theme-dot h-2.5 w-2.5 rounded-full" />
        <span class="text-lg font-bold text-white drop-shadow">{{ player.name }}</span>
      </div>

      <div class="flex-1 overflow-y-auto px-4 py-2">
        <p
          v-if="messages.length === 0"
          class="text-center text-base text-white/60"
        >
          まだメッセージはありません
        </p>
        <div
          v-else
          class="flex flex-col gap-2"
        >
          <div
            v-for="msg in messages"
            :key="msg.id"
            class="flex"
            :class="msg.senderId === myPlayerId ? 'justify-end' : 'justify-start'"
          >
            <div
              class="max-w-[70%] rounded-2xl px-4 py-2 text-base backdrop-blur-sm"
              :class="
                msg.senderId === myPlayerId
                  ? 'theme-bubble rounded-br-sm text-white'
                  : 'rounded-bl-sm bg-white/70 text-gray-900'
              "
            >
              <p>{{ msg.content }}</p>
              <p
                class="mt-0.5 text-right text-xs"
                :class="msg.senderId === myPlayerId ? 'text-white/70' : 'text-gray-500'"
              >
                {{ formatTime(msg.createdAt) }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <div class="shrink-0 px-3 pb-5 pt-2">
        <div class="flex items-center gap-2 rounded-2xl bg-white/20 px-4 py-2.5 backdrop-blur-sm">
          <p
            v-if="!canChat"
            class="flex-1 text-base text-white/50"
          >
            {{ player.name }} さんとは隣り合っていません
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
            class="theme-btn shrink-0 rounded-full px-4 py-2 text-base font-medium text-white transition-colors disabled:opacity-30"
            :disabled="!canChat || !inputText.trim()"
            @click="sendMessage"
          >
            送信
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.theme-btn,
.theme-bubble {
  background-color: rgb(var(--pc-r) var(--pc-g) var(--pc-b) / 80%);
}

.theme-btn:hover:not(:disabled) {
  background-color: rgb(var(--pc-r) var(--pc-g) var(--pc-b) / 65%);
}

.theme-dot {
  background-color: rgb(var(--pc-r) var(--pc-g) var(--pc-b));
}
</style>
