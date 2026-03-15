<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue';
import maplibregl, { type GeoJSONSource } from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import type { MapPoint } from '@/types/map';

const props = defineProps<{ points: MapPoint[] }>();

const mapContainer = ref<HTMLDivElement | null>(null);
let map: maplibregl.Map | null = null;

const POINTS_SOURCE_ID = 'points';
const LINE_SOURCE_ID = 'line';

function toPointsGeoJSON(points: MapPoint[]): GeoJSON.FeatureCollection {
  return {
    type: 'FeatureCollection',
    features: points.map(p => ({
      type: 'Feature',
      geometry: { type: 'Point', coordinates: [p.lng, p.lat] },
      properties: { id: p.id },
    })),
  };
}

function greatCircleSegment(from: MapPoint, to: MapPoint, steps = 64): [number, number][] {
  const toRad = (d: number) => (d * Math.PI) / 180;
  const toDeg = (r: number) => (r * 180) / Math.PI;
  const lat1 = toRad(from.lat), lng1 = toRad(from.lng);
  const lat2 = toRad(to.lat), lng2 = toRad(to.lng);
  const d = 2 * Math.asin(Math.sqrt(
    Math.sin((lat2 - lat1) / 2) ** 2
    + Math.cos(lat1) * Math.cos(lat2) * Math.sin((lng2 - lng1) / 2) ** 2,
  ));
  if (d === 0) return [[from.lng, from.lat]];
  return Array.from({ length: steps + 1 }, (_, i) => {
    const t = i / steps;
    const A = Math.sin((1 - t) * d) / Math.sin(d);
    const B = Math.sin(t * d) / Math.sin(d);
    const x = A * Math.cos(lat1) * Math.cos(lng1) + B * Math.cos(lat2) * Math.cos(lng2);
    const y = A * Math.cos(lat1) * Math.sin(lng1) + B * Math.cos(lat2) * Math.sin(lng2);
    const z = A * Math.sin(lat1) + B * Math.sin(lat2);
    return [toDeg(Math.atan2(y, x)), toDeg(Math.atan2(z, Math.sqrt(x * x + y * y)))];
  });
}

// 日付変更線越えで経度が飛ばないよう、前の点との差が±180°以内になるよう補正する
function unwrapLongitudes(coords: [number, number][]): [number, number][] {
  const result: [number, number][] = [];
  for (const [rawLng, lat] of coords) {
    let lng = rawLng;
    if (result.length > 0) {
      const prevLng = result[result.length - 1][0];
      while (lng - prevLng > 180) lng -= 360;
      while (prevLng - lng > 180) lng += 360;
    }
    result.push([lng, lat]);
  }
  return result;
}

function toLineGeoJSON(points: MapPoint[]): GeoJSON.FeatureCollection {
  if (points.length < 2) return { type: 'FeatureCollection', features: [] };
  const loop = [...points, points[0]];
  const coords = unwrapLongitudes(
    loop.slice(0, -1).flatMap((p, i) => greatCircleSegment(p, loop[i + 1])),
  );
  return {
    type: 'FeatureCollection',
    features: [{ type: 'Feature', geometry: { type: 'LineString', coordinates: coords }, properties: {} }],
  };
}

watch(
  () => props.points,
  (points) => {
    // getSource() の戻り値は Source 基底型で setData を持たないため、GeoJSONSource にキャストする
    (map?.getSource(POINTS_SOURCE_ID) as GeoJSONSource | undefined)?.setData(toPointsGeoJSON(points));
    (map?.getSource(LINE_SOURCE_ID) as GeoJSONSource | undefined)?.setData(toLineGeoJSON(points));
  },
);

onMounted(() => {
  if (!mapContainer.value) return;

  let minLng = Infinity, maxLng = -Infinity, minLat = Infinity, maxLat = -Infinity;
  for (const { lng, lat } of props.points) {
    if (lng < minLng) minLng = lng;
    if (lng > maxLng) maxLng = lng;
    if (lat < minLat) minLat = lat;
    if (lat > maxLat) maxLat = lat;
  }
  const hasBounds = props.points.length > 0;

  map = new maplibregl.Map({
    container: mapContainer.value,
    style: 'https://tiles.openfreemap.org/styles/liberty',
    bounds: hasBounds ? new maplibregl.LngLatBounds([minLng, minLat], [maxLng, maxLat]) : undefined,
    fitBoundsOptions: { padding: 80 },
    center: hasBounds ? undefined : [131.467, 33.337],
    zoom: hasBounds ? undefined : 15,
  });

  map.addControl(new maplibregl.NavigationControl(), 'top-right');

  map.on('load', () => {
    const m = map!;

    for (const layer of m.getStyle().layers) {
      if (layer.type === 'symbol') {
        m.setLayoutProperty(layer.id, 'visibility', 'none');
      }
    }

    m.addSource(LINE_SOURCE_ID, { type: 'geojson', data: toLineGeoJSON(props.points) });
    m.addSource(POINTS_SOURCE_ID, { type: 'geojson', data: toPointsGeoJSON(props.points) });

    m.addLayer({
      id: 'line',
      type: 'line',
      source: LINE_SOURCE_ID,
      paint: {
        'line-color': '#FF4444',
        'line-width': 2,
        'line-opacity': 0.8,
      },
    });

    m.addLayer({
      id: 'points-circle',
      type: 'circle',
      source: POINTS_SOURCE_ID,
      paint: {
        'circle-radius': 10,
        'circle-color': '#FF4444',
        'circle-stroke-width': 3,
        'circle-stroke-color': '#FFFFFF',
      },
    });
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
