<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import GameScene from '@/components/GameScene.vue';
import MatchingScene from '@/components/MatchingScene.vue';
import type { Player, Message } from '@/types/game';

const route = useRoute();
const router = useRouter();

const roomId = route.query.room_id as string;
const playerId = route.query.player_id as string;

const roomStatus = ref<'matching' | 'playing' | 'finished'>('matching');
const players = ref<Player[]>([]);
const messages = ref<Message[]>([]);

type ApiPlayer = {
  id: string;
  name: string;
  order_index: number;
  location: { lat: number; lng: number };
};

function mapPlayer(p: ApiPlayer): Player {
  return {
    id: p.id,
    name: p.name,
    orderIndex: p.order_index,
    lat: p.location.lat,
    lng: p.location.lng,
  };
}

async function fetchRoomState() {
  const res = await fetch(`/api/rooms/${roomId}`, {
    headers: { 'X-Player-ID': playerId },
  });
  if (!res.ok) return;
  const data = await res.json() as {
    status: 'matching' | 'playing' | 'finished';
    players: ApiPlayer[];
  };
  roomStatus.value = data.status;
  players.value = data.players.map(mapPlayer);
}

const matchingPoints = computed(() =>
  players.value.map(p => ({ id: p.id, lat: p.lat, lng: p.lng, name: p.name })),
);

let sse: EventSource | null = null;

onMounted(async () => {
  if (!roomId || !playerId) {
    router.replace('/');
    return;
  }
  await fetchRoomState();
  sse = new EventSource(`/api/rooms/${roomId}/stream?player_id=${playerId}`);
  sse.addEventListener('room_started', () => {
    void fetchRoomState();
  });
  sse.addEventListener('room_updated', () => {
    void fetchRoomState();
  });
});

onUnmounted(() => {
  sse?.close();
});

function onSendMessage(receiverId: string, content: string) {
  // TODO: POST /api/rooms/{room_id}/messages に接続する
  messages.value.push({
    id: crypto.randomUUID(),
    senderId: playerId,
    receiverId,
    content,
    createdAt: new Date(),
  });
}

function onToggleSwap(targetPlayerId: string, needsSwap: boolean) {
  // TODO: PUT /api/rooms/{room_id}/connections/{target_id} に接続する
  console.log('[GamePage] toggleSwap', targetPlayerId, needsSwap);
}
</script>

<template>
  <DefaultLayout>
    <div class="h-svh">
      <GameScene
        v-if="roomStatus !== 'matching'"
        :my-player-id="playerId"
        :players="players"
        :messages="messages"
        @send-message="onSendMessage"
        @toggle-swap="onToggleSwap"
      />
      <MatchingScene
        v-else
        :my-player-id="playerId"
        :max-count="4"
        :points="matchingPoints"
      />
    </div>
  </DefaultLayout>
</template>
