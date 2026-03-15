<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue';
import maplibregl, { type GeoJSONSource } from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import type { MapPoint } from '@/types/map';

const props = defineProps<{
  points: MapPoint[];
  highlightedId?: string;
}>();

const mapContainer = ref<HTMLDivElement | null>(null);
let map: maplibregl.Map | null = null;

const POINTS_SOURCE_ID = 'points';
const LINE_SOURCE_ID = 'line';
type Rgb = [number, number, number];
type Role = 'highlight' | 'adjacent' | 'normal';

// ID を HSL の RGB に変換する
// 黄金角 (137.508°) を掛けることで、連番 ID でも hue が均等に散らばる
// role に応じて彩度・明度を調整する: highlight は鮮やか、adjacent は通常、normal は淡め
function idToRgb(id: string, role: Role): Rgb {
  const n = [...id].reduce((h, c) => (h * 31 + c.charCodeAt(0)) | 0, 0);
  const h = Math.abs(n) * 137.508 % 360;
  const [s, l] = role === 'highlight'
    ? [0.90, 0.48]
    : role === 'adjacent'
      ? [0.75, 0.55]
      : [0.45, 0.65];
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

// highlightedId に隣接している (ループ上で隣の) 点の ID セットを返す
function adjacentIdsOf(points: MapPoint[], highlightedId: string | undefined): Set<string> {
  if (!highlightedId) return new Set();
  const idx = points.findIndex(p => p.id === highlightedId);
  if (idx === -1) return new Set();
  return new Set([
    points[(idx - 1 + points.length) % points.length].id,
    points[(idx + 1) % points.length].id,
  ]);
}

function toPointsGeoJSON(points: MapPoint[], highlightedId: string | undefined): GeoJSON.FeatureCollection {
  const adjIds = adjacentIdsOf(points, highlightedId);
  return {
    type: 'FeatureCollection',
    features: points.map((p) => {
      const role: Role = p.id === highlightedId
        ? 'highlight'
        : adjIds.has(p.id)
          ? 'adjacent'
          : 'normal';
      const [r, g, b] = idToRgb(p.id, role);
      return {
        type: 'Feature',
        geometry: { type: 'Point', coordinates: [p.lng, p.lat] },
        properties: { id: p.id, r, g, b, role },
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
function toLineGeoJSON(points: MapPoint[], highlightedId: string | undefined): GeoJSON.FeatureCollection {
  if (points.length < 2) return { type: 'FeatureCollection', features: [] };
  const loop = [...points, points[0]];
  const features: GeoJSON.Feature[] = [];

  for (let i = 0; i < loop.length - 1; i++) {
    const role: Role = (highlightedId && (loop[i].id === highlightedId || loop[i + 1].id === highlightedId))
      ? 'adjacent'
      : 'normal';
    const rgbFrom = idToRgb(loop[i].id, role);
    const rgbTo = idToRgb(loop[i + 1].id, role);
    const coords = unwrapLongitudes(greatCircleSegment(loop[i], loop[i + 1]));

    for (let j = 0; j < coords.length - 1; j++) {
      const [r, g, b] = lerpRgb(rgbFrom, rgbTo, j / (coords.length - 1));
      features.push({
        type: 'Feature',
        geometry: { type: 'LineString', coordinates: [coords[j], coords[j + 1]] },
        properties: { r, g, b, role },
      });
    }
  }

  return { type: 'FeatureCollection', features };
}

watch(
  [() => props.points, () => props.highlightedId],
  ([points, highlightedId]) => {
    // getSource() の戻り値は Source 基底型で setData を持たないため、GeoJSONSource にキャストする
    (map?.getSource(POINTS_SOURCE_ID) as GeoJSONSource | undefined)?.setData(toPointsGeoJSON(points, highlightedId));
    (map?.getSource(LINE_SOURCE_ID) as GeoJSONSource | undefined)?.setData(toLineGeoJSON(points, highlightedId));
  },
);

onMounted(() => {
  if (!mapContainer.value) return;

  // TODO: 日付変更線をまたぐ場合の対応をする (ハッカソン中は放置)
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

    m.addSource(LINE_SOURCE_ID, { type: 'geojson', data: toLineGeoJSON(props.points, props.highlightedId) });
    m.addSource(POINTS_SOURCE_ID, { type: 'geojson', data: toPointsGeoJSON(props.points, props.highlightedId) });

    m.addLayer({
      id: 'line-outline',
      type: 'line',
      source: LINE_SOURCE_ID,
      paint: {
        'line-width': ['match', ['get', 'role'], 'adjacent', 7, 4],
        'line-color': '#111111',
        'line-opacity': ['match', ['get', 'role'], 'adjacent', 0.9, 0.4],
      },
    });

    m.addLayer({
      id: 'line',
      type: 'line',
      source: LINE_SOURCE_ID,
      paint: {
        'line-width': ['match', ['get', 'role'], 'adjacent', 4, 1.5],
        'line-color': ['rgb', ['get', 'r'], ['get', 'g'], ['get', 'b']],
        'line-opacity': ['match', ['get', 'role'], 'adjacent', 1.0, 0.55],
      },
    });

    m.addLayer({
      id: 'points-circle',
      type: 'circle',
      source: POINTS_SOURCE_ID,
      paint: {
        'circle-radius': ['match', ['get', 'role'], 'highlight', 14, 'adjacent', 10, 7],
        'circle-color': ['rgb', ['get', 'r'], ['get', 'g'], ['get', 'b']],
        'circle-stroke-width': ['match', ['get', 'role'], 'highlight', 4, 'adjacent', 3, 2],
        'circle-stroke-color': '#111111',
        'circle-opacity': ['match', ['get', 'role'], 'normal', 0.65, 1.0],
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
