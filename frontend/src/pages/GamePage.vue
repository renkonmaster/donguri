<script setup lang="ts">
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import GameMap from '@/components/GameMap.vue';
import type { MapPoint } from '@/types/map';

const points: MapPoint[] = [
  { id: '1', lat: -33.868, lng: 151.209, name: 'Alice' },
  { id: '2', lat: 40.712, lng: -74.006, name: 'Bob' },
  { id: '3', lat: 55.751, lng: 37.617, name: 'Carol' },
  { id: '4', lat: -23.550, lng: -46.633, name: 'Dave' },
  { id: '5', lat: 35.689, lng: 139.692, name: 'りすりす' },
  { id: '6', lat: 1.352, lng: 103.820, name: 'Frank' },
  { id: '7', lat: 51.507, lng: -0.127, name: 'Grace' },
  { id: '8', lat: 64.135, lng: -21.895, name: 'Hank' },
  { id: '9', lat: 37.774, lng: -122.419, name: 'Ivy' },
  { id: '10', lat: -1.286, lng: 36.817, name: 'Jack' },
  { id: '11', lat: 19.432, lng: -99.133, name: 'Kate' },
  { id: '12', lat: 59.913, lng: 10.752, name: 'Leo' },
  { id: '13', lat: 31.235, lng: 121.473, name: 'Mia' },
  { id: '14', lat: -34.603, lng: -58.381, name: 'Nick' },
  { id: '15', lat: 28.613, lng: 77.209, name: 'Olivia' },
];
</script>

<template>
  <DefaultLayout>
    <div class="h-[70svh] w-full overflow-hidden rounded-xl">
      <GameMap
        :points="points"
        highlighted-id="5"
        show-line
      />
    </div>
  </DefaultLayout>
</template>
