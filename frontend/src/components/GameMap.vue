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
type Rgb = [number, number, number];

// ID を HSL(h, 80%, 55%) の RGB に変換する
// 黄金角 (137.508°) を掛けることで、連番 ID でも hue が均等に散らばる
function idToRgb(id: string): Rgb {
  const n = [...id].reduce((h, c) => (h * 31 + c.charCodeAt(0)) | 0, 0);
  const h = Math.abs(n) * 137.508 % 360;
  const s = 0.8, l = 0.55;
  const a = s * Math.min(l, 1 - l);
  const channel = (pos: number) => {
    const k = (pos + h / 30) % 12;
    return Math.round((l - a * Math.max(-1, Math.min(k - 3, Math.min(9 - k, 1)))) * 255);
  };
  return [channel(0), channel(8), channel(4)];
}

function lerpRgb(a: Rgb, b: Rgb, t: number): Rgb {
  return [
    Math.round(a[0] + (b[0] - a[0]) * t),
    Math.round(a[1] + (b[1] - a[1]) * t),
    Math.round(a[2] + (b[2] - a[2]) * t),
  ];
}

function toPointsGeoJSON(points: MapPoint[]): GeoJSON.FeatureCollection {
  return {
    type: 'FeatureCollection',
    features: points.map((p) => {
      const [r, g, b] = idToRgb(p.id);
      return {
        type: 'Feature',
        geometry: { type: 'Point', coordinates: [p.lng, p.lat] },
        properties: { id: p.id, r, g, b },
      };
    }),
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

// greatCircleSegment の隣接 2 点を 1 Feature とし、両端色の補間色を付与してグラデーションを表現する
function toLineGeoJSON(points: MapPoint[]): GeoJSON.FeatureCollection {
  if (points.length < 2) return { type: 'FeatureCollection', features: [] };
  const loop = [...points, points[0]];
  const features: GeoJSON.Feature[] = [];

  for (let i = 0; i < loop.length - 1; i++) {
    const rgbFrom = idToRgb(loop[i].id);
    const rgbTo = idToRgb(loop[i + 1].id);
    const coords = unwrapLongitudes(greatCircleSegment(loop[i], loop[i + 1]));

    for (let j = 0; j < coords.length - 1; j++) {
      const [r, g, b] = lerpRgb(rgbFrom, rgbTo, j / (coords.length - 1));
      features.push({
        type: 'Feature',
        geometry: { type: 'LineString', coordinates: [coords[j], coords[j + 1]] },
        properties: { r, g, b },
      });
    }
  }

  return { type: 'FeatureCollection', features };
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
      id: 'line-outline',
      type: 'line',
      source: LINE_SOURCE_ID,
      paint: {
        'line-width': 5,
        'line-color': '#111111',
      },
    });

    m.addLayer({
      id: 'line',
      type: 'line',
      source: LINE_SOURCE_ID,
      paint: {
        'line-width': 3,
        'line-color': ['rgb', ['get', 'r'], ['get', 'g'], ['get', 'b']],
      },
    });

    m.addLayer({
      id: 'points-circle',
      type: 'circle',
      source: POINTS_SOURCE_ID,
      paint: {
        'circle-radius': 10,
        'circle-color': ['rgb', ['get', 'r'], ['get', 'g'], ['get', 'b']],
        'circle-stroke-width': 3,
        'circle-stroke-color': '#111111',
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
