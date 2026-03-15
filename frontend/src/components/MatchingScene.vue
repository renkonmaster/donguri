<script setup lang="ts">
import GameMap from '@/components/GameMap.vue';
import type { MapPoint } from '@/types/map';

const props = defineProps<{
  myPlayerId: string;
  maxCount: number;
  points: MapPoint[];
  countdownSeconds?: number | null;
}>();
</script>

<template>
  <div class="flex h-full flex-col">
    <div class="flex-1 overflow-hidden">
      <GameMap
        :points="props.points"
        :highlighted-id="props.myPlayerId"
        :show-line="false"
        auto-fit
      />
    </div>

    <div class="relative border-t border-gray-200 bg-white px-4 py-3">
      <span class="block text-center text-lg font-medium text-gray-700">
        <template v-if="props.countdownSeconds != null">
          ゲームまであと {{ props.countdownSeconds }} 秒…
        </template>
        <template v-else>
          {{ props.points.length }}/{{ props.maxCount }} 人参加
        </template>
      </span>

      <!-- ハッカソン中は退出機能未実装のため常にグレーアウト -->
      <button
        disabled
        class="absolute right-4 top-1/2 -translate-y-1/2 cursor-not-allowed rounded-lg bg-gray-200 px-4 py-2 text-sm font-medium text-gray-400"
      >
        退出
      </button>
    </div>
  </div>
</template>
