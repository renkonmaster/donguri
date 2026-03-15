<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
import maplibregl from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';

const mapContainer = ref<HTMLDivElement | null>(null);
let map: maplibregl.Map | null = null;

onMounted(() => {
  if (!mapContainer.value) return;

  map = new maplibregl.Map({
    container: mapContainer.value,
    style: 'https://tiles.openfreemap.org/styles/liberty',
    center: [135.5, 34.7],
    zoom: 13,
  });

  map.addControl(new maplibregl.NavigationControl(), 'top-right');

  map.on('load', () => {
    for (const layer of map!.getStyle().layers) {
      if (layer.type === 'symbol') {
        map!.setLayoutProperty(layer.id, 'visibility', 'none');
      }
    }
  });
});

onUnmounted(() => {
  map?.remove();
  map = null;
});
</script>

<template>
  <div
    ref="mapContainer"
    class="size-full"
  />
</template>
