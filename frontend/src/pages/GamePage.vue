<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRoute } from 'vue-router';
import { useOgpHead } from '@/composables/useOgpHead';
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import GameScene from '@/components/GameScene.vue';
import MatchingScene from '@/components/MatchingScene.vue';
import type { Player, Message } from '@/types/game';

type Scene = 'matching' | 'game';

// デバッグ用: URL ハッシュでシーンを切り替え (/game#matching → matching, /game → game)
const route = useRoute();
const scene = computed<Scene>(() => route.hash === '#matching' ? 'matching' : 'game');

useOgpHead(
  'InterKnot | ゲームページ',
  'InterKnotのゲームページです。マッチング状況や参加者の位置情報を確認できます。',
);

const myPlayerId = 'p5';

const players: Player[] = [
  { id: 'p1', name: 'Alice', orderIndex: 0, lat: -33.868, lng: 151.209 },
  { id: 'p2', name: 'Bob', orderIndex: 1, lat: 40.712, lng: -74.006 },
  { id: 'p3', name: 'Carol', orderIndex: 2, lat: 55.751, lng: 37.617 },
  { id: 'p4', name: 'Dave', orderIndex: 3, lat: -23.550, lng: -46.633 },
  { id: 'p5', name: 'りすりす', orderIndex: 4, lat: 35.689, lng: 139.692 },
  { id: 'p6', name: 'Frank', orderIndex: 5, lat: 1.352, lng: 103.820 },
];

const messages = ref<Message[]>([
  {
    id: 'm1',
    senderId: 'p4',
    receiverId: 'p5',
    content: 'ねえ、交換しようよ！',
    createdAt: new Date('2025-01-01T12:00:00'),
  },
  {
    id: 'm2',
    senderId: 'p5',
    receiverId: 'p4',
    content: 'いいよ、交換ボタン押すね',
    createdAt: new Date('2025-01-01T12:00:30'),
  },
]);

function onSendMessage(receiverId: string, content: string) {
  messages.value.push({
    id: crypto.randomUUID(),
    senderId: myPlayerId,
    receiverId,
    content,
    createdAt: new Date(),
  });
}

function onToggleSwap(targetPlayerId: string, needsSwap: boolean) {
  console.log('[GamePage] toggleSwap', targetPlayerId, needsSwap);
}

// マッチング画面用スタブ
const matchingMaxCount = 6;
const matchingPoints = [players[0], players[2], players[3], players.find(p => p.id === myPlayerId)!].map(p => ({
  id: p.id,
  lat: p.lat,
  lng: p.lng,
  name: p.name,
}));
</script>

<template>
  <DefaultLayout>
    <div class="h-svh">
      <GameScene
        v-if="scene === 'game'"
        :my-player-id="myPlayerId"
        :players="players"
        :messages="messages"
        @send-message="onSendMessage"
        @toggle-swap="onToggleSwap"
      />
      <MatchingScene
        v-else-if="scene === 'matching'"
        :my-player-id="myPlayerId"
        :max-count="matchingMaxCount"
        :points="matchingPoints"
      />
    </div>
  </DefaultLayout>
</template>
