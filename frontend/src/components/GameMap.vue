<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue';
import type { GeoJSONSource, Map as MaplibreMap } from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import type { MapClickPayload, MapPoint } from '@/types/map';
import { idToRgb, lerpRgb, type Role } from '@/utils/pointColor';
import { greatCircleSegment, unwrapLongitudes } from '@/utils/geo';

const props = defineProps<{
  points: MapPoint[];
  highlightedId?: string;
  showLine?: boolean;
  autoFit?: boolean;
  unreadIds?: string[];
}>();

const emit = defineEmits<{
  click: [payload: MapClickPayload];
}>();

const mapContainer = ref<HTMLDivElement | null>(null);
let map: MaplibreMap | null = null;

function calcBounds(points: MapPoint[]) {
  let minLng = Infinity, maxLng = -Infinity, minLat = Infinity, maxLat = -Infinity;
  for (const { lng, lat } of points) {
    if (lng < minLng) minLng = lng;
    if (lng > maxLng) maxLng = lng;
    if (lat < minLat) minLat = lat;
    if (lat > maxLat) maxLat = lat;
  }
  return { minLng, maxLng, minLat, maxLat };
}

const POINTS_SOURCE_ID = 'points';
const LINE_SOURCE_ID = 'line';
const LAYER_LINE_OUTLINE = 'line-outline';
const LAYER_LINE = 'line-color';
const LAYER_POINTS_CIRCLE = 'points-circle';
const LAYER_POINTS_LABEL = 'points-label';
const LAYER_POINTS_HITAREA = 'points-hitarea';

// 未読パルス位置: map.project() でプライマリ画素座標を取り、worldSize 分ずらして 3 コピー分を Vue で描画
type UnreadPos = { key: string; x: number; y: number; r: number; g: number; b: number };
const unreadPositions = ref<UnreadPos[]>([]);

function updateUnreadPositions() {
  if (!map) {
    unreadPositions.value = [];
    return;
  }
  const worldSize = 512 * Math.pow(2, map.getZoom());
  const positions: UnreadPos[] = [];
  for (const id of (props.unreadIds ?? [])) {
    const p = props.points.find(pt => pt.id === id);
    if (!p) continue;
    const [r, g, b] = idToRgb(p.id, 'highlight');
    const primary = map.project([p.lng, p.lat]);
    for (const n of [-1, 0, 1]) {
      positions.push({ key: `${id}:${n}`, x: primary.x + n * worldSize, y: primary.y, r, g, b });
    }
  }
  unreadPositions.value = positions;
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

function toPointsGeoJSON(
  points: MapPoint[],
  highlightedId: string | undefined,
  showAdjacency: boolean,
): GeoJSON.FeatureCollection {
  const adjacentIds = showAdjacency ? adjacentIdsOf(points, highlightedId) : new Set<string>();
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

// 各辺を greatCircleSegment で補間し、補間した各セグメントを 1 Feature として両端色の補間色を付与する
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
      // coords.length > 2 のとき j / (coords.length - 2) で最終セグメントが t=1 (終点色) になる
      // coords.length === 2 は 1 本のみで分母がゼロになるため t=0 (始点色) とする
      const t = coords.length > 2 ? j / (coords.length - 2) : 0;
      const [r, g, b] = lerpRgb(rgbFrom, rgbTo, t);
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
    // load イベント発火前はレイヤーが存在しないため、存在確認してから操作する
    if (!map?.getLayer(LAYER_LINE_OUTLINE)) return;
    const visibility = (showLine ?? true) ? 'visible' : 'none';
    map.setLayoutProperty(LAYER_LINE_OUTLINE, 'visibility', visibility);
    map.setLayoutProperty(LAYER_LINE, 'visibility', visibility);
  },
);

// 注意: deep: true を付けていないため、親は points を in-place で変更せず
// 必ず新しい配列を渡すこと (immutable update)。
// deep watch にするとポイント数に比例してコストが増えるため採用しない。
watch(
  [() => props.points, () => props.highlightedId],
  ([points, highlightedId]) => {
    // getSource() の戻り値は Source 基底型で setData を持たないため、GeoJSONSource にキャストする
    const showAdjacency = props.showLine ?? true;
    const pointsSrc = map?.getSource(POINTS_SOURCE_ID) as GeoJSONSource | undefined;
    const lineSrc = map?.getSource(LINE_SOURCE_ID) as GeoJSONSource | undefined;
    pointsSrc?.setData(toPointsGeoJSON(points, highlightedId, showAdjacency));
    lineSrc?.setData(toLineGeoJSON(points, highlightedId));
    updateUnreadPositions();
    if (props.autoFit && map && points.length > 0) {
      const { minLng, maxLng, minLat, maxLat } = calcBounds(points);
      map.fitBounds([[minLng, minLat], [maxLng, maxLat]], { padding: 80 });
    }
  },
);

watch(() => props.unreadIds, updateUnreadPositions);

onMounted(async () => {
  if (!mapContainer.value) return;

  // onMounted はブラウザ環境のみで実行されるため、SSR クラッシュを防ぐために動的 import する
  const maplibregl = (await import('maplibre-gl')).default;

  // TODO: 日付変更線をまたぐ場合の対応をする (ハッカソン中は放置)
  const hasBounds = props.points.length > 0;
  const { minLng, maxLng, minLat, maxLat } = calcBounds(props.points);

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
    const initShowAdjacency = props.showLine ?? true;
    const initPoints = toPointsGeoJSON(props.points, props.highlightedId, initShowAdjacency);
    m.addSource(POINTS_SOURCE_ID, { type: 'geojson', data: initPoints });

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

    m.on('move', updateUnreadPositions);
    updateUnreadPositions();

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
  <div class="relative size-full">
    <div
      ref="mapContainer"
      class="size-full"
    />
    <!-- 未読パルスオーバーレイ: CSS アニメーション + ワールドコピー対応 -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div
        v-for="pos in unreadPositions"
        :key="pos.key"
        class="unread-pulse-wrapper"
        :style="{ '--pr': pos.r, '--pg': pos.g, '--pb': pos.b, transform: `translate(${pos.x}px, ${pos.y}px)` }"
      >
        <div class="unread-pulse-ring" />
      </div>
    </div>
  </div>
</template>

<style scoped>
@keyframes unread-ping {
  from {
    transform: scale(1);
    opacity: 0.8;
  }

  to {
    transform: scale(3.5);
    opacity: 0;
  }
}

.unread-pulse-wrapper {
  position: absolute;
  top: 0;
  left: 0;
  width: 0;
  height: 0;
}

.unread-pulse-ring {
  position: absolute;
  top: -14px;
  left: -14px;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background-color: rgb(var(--pr) var(--pg) var(--pb));
  animation: unread-ping 0.5s linear infinite;
  pointer-events: none;
}
</style>
