<script setup lang="ts">
import { ref, reactive, computed } from 'vue';
import GameMap from '@/components/GameMap.vue';
import ChatOverlay from '@/components/ChatOverlay.vue';
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
const swapRequested = reactive(new Set<string>());

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
</style>
