<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue';
import maplibregl, { type GeoJSONSource } from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import type { MapClickPayload, MapPoint } from '@/types/map';
import { idToRgb, lerpRgb, type Role } from '@/utils/pointColor';
import { greatCircleSegment, unwrapLongitudes } from '@/utils/geo';

const props = defineProps<{
  points: MapPoint[];
  highlightedId?: string;
  showLine?: boolean;
}>();

const emit = defineEmits<{
  click: [payload: MapClickPayload];
}>();

const mapContainer = ref<HTMLDivElement | null>(null);
let map: maplibregl.Map | null = null;

const POINTS_SOURCE_ID = 'points';
const LINE_SOURCE_ID = 'line';
const LAYER_LINE_OUTLINE = 'line-outline';
const LAYER_LINE = 'line';
const LAYER_POINTS_CIRCLE = 'points-circle';
const LAYER_POINTS_LABEL = 'points-label';
const LAYER_POINTS_HITAREA = 'points-hitarea';

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
  const adjacentIds = adjacentIdsOf(points, highlightedId);
  return {
    type: 'FeatureCollection',
    features: points.map((p) => {
      const role: Role = p.id === highlightedId
        ? 'highlight'
        : adjacentIds.has(p.id)
          ? 'adjacent'
          : 'normal';
      const [r, g, b] = idToRgb(p.id, role);
      return {
        type: 'Feature',
        geometry: { type: 'Point', coordinates: [p.lng, p.lat] },
        properties: { id: p.id, r, g, b, role, name: p.name ?? '' },
      };
    }),
  };
}

// 各辺を greatCircleSegment で補間し、隣接 2 点を 1 Feature として両端色の補間色を付与する
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
  () => props.showLine,
  (showLine) => {
    const visibility = (showLine ?? true) ? 'visible' : 'none';
    map?.setLayoutProperty(LAYER_LINE_OUTLINE, 'visibility', visibility);
    map?.setLayoutProperty(LAYER_LINE, 'visibility', visibility);
  },
);

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
      id: LAYER_LINE_OUTLINE,
      type: 'line',
      source: LINE_SOURCE_ID,
      paint: {
        'line-width': ['match', ['get', 'role'], 'adjacent', 7, 4],
        'line-color': '#111111',
        'line-opacity': ['match', ['get', 'role'], 'adjacent', 0.9, 0.4],
      },
    });

    m.addLayer({
      id: LAYER_LINE,
      type: 'line',
      source: LINE_SOURCE_ID,
      paint: {
        'line-width': ['match', ['get', 'role'], 'adjacent', 4, 1.5],
        'line-color': ['rgb', ['get', 'r'], ['get', 'g'], ['get', 'b']],
        'line-opacity': ['match', ['get', 'role'], 'adjacent', 1.0, 0.55],
      },
    });

    const lineVisibility = (props.showLine ?? true) ? 'visible' : 'none';
    m.setLayoutProperty(LAYER_LINE_OUTLINE, 'visibility', lineVisibility);
    m.setLayoutProperty(LAYER_LINE, 'visibility', lineVisibility);

    m.addLayer({
      id: LAYER_POINTS_CIRCLE,
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

    m.addLayer({
      id: LAYER_POINTS_LABEL,
      type: 'symbol',
      source: POINTS_SOURCE_ID,
      layout: {
        'text-field': ['get', 'name'],
        'text-font': ['Noto Sans Bold', 'Noto Sans Regular'],
        'text-anchor': 'bottom',
        'text-offset': [0, -1.0],
        'text-size': ['match', ['get', 'role'], 'highlight', 20, 'adjacent', 17, 15],
      },
      paint: {
        'text-color': ['rgb', ['get', 'r'], ['get', 'g'], ['get', 'b']],
        'text-halo-color': '#111111',
        'text-halo-width': ['match', ['get', 'role'], 'highlight', 2, 1],
      },
    });

    // 視覚的な円より大きい当たり判定用の透明レイヤー
    m.addLayer({
      id: LAYER_POINTS_HITAREA,
      type: 'circle',
      source: POINTS_SOURCE_ID,
      paint: {
        'circle-radius': 48,
        'circle-opacity': 0,
      },
    });

    m.on('click', (e) => {
      const pointById = new Map(props.points.map(p => [p.id, p]));
      const { lat, lng } = e.lngLat;
      const candidates = m.queryRenderedFeatures(e.point, { layers: [LAYER_POINTS_HITAREA] })
        .map(f => pointById.get(f.properties?.id))
        .filter((p): p is MapPoint => p !== undefined);
      const point = candidates.length > 0
        ? candidates
          .map(p => ({ p, d: (p.lat - lat) ** 2 + (p.lng - lng) ** 2 }))
          .reduce((a, b) => a.d < b.d ? a : b)
          .p
        : undefined;
      emit('click', { lat, lng, point });
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
