<script setup lang="ts">
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import GameMap from '@/components/GameMap.vue';
import type { MapPoint } from '@/types/map';

const points: MapPoint[] = [
  { id: '1', lat: -33.868, lng: 151.209 }, // シドニー
  { id: '2', lat: 40.712, lng: -74.006 }, // ニューヨーク
  { id: '3', lat: 55.751, lng: 37.617 }, // モスクワ
  { id: '4', lat: -23.550, lng: -46.633 }, // サンパウロ
  { id: '5', lat: 35.689, lng: 139.692 }, // 東京
  { id: '6', lat: 1.352, lng: 103.820 }, // シンガポール
  { id: '7', lat: 51.507, lng: -0.127 }, // ロンドン
  { id: '8', lat: 64.135, lng: -21.895 }, // レイキャビク
  { id: '9', lat: 37.774, lng: -122.419 }, // サンフランシスコ
  { id: '10', lat: -1.286, lng: 36.817 }, // ナイロビ
  { id: '11', lat: 19.432, lng: -99.133 }, // メキシコシティ
  { id: '12', lat: 59.913, lng: 10.752 }, // オスロ
  { id: '13', lat: 31.235, lng: 121.473 }, // 上海
  { id: '14', lat: -34.603, lng: -58.381 }, // ブエノスアイレス
  { id: '15', lat: 28.613, lng: 77.209 }, // デリー
];
</script>

<template>
  <DefaultLayout>
    <div class="h-[70svh] w-full overflow-hidden rounded-xl">
      <GameMap
        :points="points"
        highlighted-id="5"
      />
    </div>
  </DefaultLayout>
</template>
