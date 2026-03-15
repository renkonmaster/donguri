<script setup lang="ts">
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import GameMap from '@/components/GameMap.vue';
</script>

<template>
  <DefaultLayout>
    <div class="h-[70svh] w-full overflow-hidden rounded-xl">
      <GameMap />
    </div>
  </DefaultLayout>
</template>
