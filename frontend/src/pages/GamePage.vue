<script setup lang="ts">
import MatchingScene from '@/components/MatchingScene.vue';
import type { MapPoint } from '@/types/map';

// 実験用ダミーデータ
const points: MapPoint[] = [
  { id: '1', lat: 35.689, lng: 139.692, name: 'Alice' },
  { id: '2', lat: 35.694, lng: 139.700, name: 'Bob' },
  { id: '3', lat: 35.683, lng: 139.685, name: 'Carol' },
];
</script>

<template>
  <div class="h-svh">
    <MatchingScene
      :current-count="3"
      :max-count="10"
      :points="points"
    />
  </div>
</template>
