<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useOgpHead } from '@/composables/useOgpHead';
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import GameScene from '@/components/GameScene.vue';
import MatchingScene from '@/components/MatchingScene.vue';
import type { Player, Message } from '@/types/game';

const route = useRoute();
const router = useRouter();

useOgpHead(
  'InterKnot | ゲームページ',
  'InterKnotのゲームページです。マッチング状況や参加者の位置情報を確認できます。',
);

const roomId = typeof route.query.room_id === 'string' ? route.query.room_id : '';
const playerId = typeof route.query.player_id === 'string' ? route.query.player_id : '';

const roomStatus = ref<'matching' | 'playing' | 'finished'>('matching');
const players = ref<Player[]>([]);
const messages = ref<Message[]>([]);
const countdownSeconds = ref<number | null>(null);
let countdownTimer: ReturnType<typeof setInterval> | null = null;

const gameStartTime = ref<number | null>(null);
const swapCount = ref(0);
const clearTimeMs = ref<number | null>(null);

watch(roomStatus, (newStatus) => {
  if (newStatus === 'playing' && gameStartTime.value === null) {
    gameStartTime.value = Date.now();
  }
  if (newStatus === 'finished' && clearTimeMs.value === null) {
    clearTimeMs.value = gameStartTime.value !== null ? Date.now() - gameStartTime.value : null;
  }
}, { flush: 'sync' });

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

const maxPlayersPerRoom = 8;
const gameStartCountdownSeconds = 5;

async function fetchRoomState() {
  const res = await fetch(`/api/rooms/${roomId}`, {
    headers: { 'X-Player-ID': playerId },
  });
  if (res.status === 404) {
    router.replace('/');
    return;
  }
  if (!res.ok) return;
  const data = await res.json() as {
    status: 'matching' | 'playing' | 'finished';
    players: ApiPlayer[];
  };
  roomStatus.value = data.status;
  players.value = data.players.map(mapPlayer);

  if (data.status === 'matching' && data.players.length === maxPlayersPerRoom && countdownTimer === null) {
    startCountdown();
  }
}

function startCountdown() {
  countdownSeconds.value = gameStartCountdownSeconds;
  countdownTimer = setInterval(() => {
    if (countdownSeconds.value === null || countdownSeconds.value <= 1) {
      clearInterval(countdownTimer!);
      countdownTimer = null;
      countdownSeconds.value = null;
      return;
    }
    countdownSeconds.value--;
  }, 1000);
}

const matchingPoints = computed(() =>
  players.value.map(p => ({ id: p.id, orderIndex: p.orderIndex, lat: p.lat, lng: p.lng, name: p.name })),
);

let sse: EventSource | null = null;

onMounted(async () => {
  if (!roomId || !playerId) {
    router.replace('/');
    return;
  }
  await fetchRoomState();
  sse = new EventSource(`/api/rooms/${roomId}/stream?player_id=${playerId}`);
  sse.addEventListener('message_received', (e: MessageEvent) => {
    const data = JSON.parse(e.data) as unknown as {
      id: string;
      sender_id: string;
      receiver_id: string;
      content: string;
      created_at: string;
    };
    messages.value.push({
      id: data.id,
      senderId: data.sender_id,
      receiverId: data.receiver_id,
      content: data.content,
      createdAt: new Date(data.created_at),
    });
  });
  sse.addEventListener('room_started', () => {
    if (countdownTimer !== null) {
      clearInterval(countdownTimer);
      countdownTimer = null;
    }
    countdownSeconds.value = null;
    void fetchRoomState();
  });
  sse.addEventListener('room_updated', () => {
    void fetchRoomState();
  });
});

onUnmounted(() => {
  sse?.close();
  if (countdownTimer !== null) {
    clearInterval(countdownTimer);
  }
});

async function onSendMessage(receiverId: string, content: string) {
  const res = await fetch(`/api/rooms/${roomId}/messages`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Player-ID': playerId,
    },
    body: JSON.stringify({ receiver_id: receiverId, content }),
  });
  if (!res.ok) return;
  // メッセージの追加は SSE の message_received イベントで行われる
}

async function onToggleSwap(targetPlayerId: string, needsSwap: boolean) {
  const res = await fetch(`/api/rooms/${roomId}/connections/${targetPlayerId}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      'X-Player-ID': playerId,
    },
    body: JSON.stringify({ needs_swap: needsSwap }),
  });
  if (!res.ok) return;
  const data = await res.json() as unknown as { matched: boolean };
  if (data.matched) {
    swapCount.value++;
    // ルーム状態の更新は SSE の room_updated イベント経由で行われる
  }
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
        :room-status="roomStatus"
        :swap-count="swapCount"
        :clear-time-ms="clearTimeMs"
        @send-message="onSendMessage"
        @toggle-swap="onToggleSwap"
      />
      <MatchingScene
        v-else
        :my-player-id="playerId"
        :max-count="maxPlayersPerRoom"
        :points="matchingPoints"
        :countdown-seconds="countdownSeconds"
      />
    </div>
  </DefaultLayout>
</template>
