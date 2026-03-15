<script setup lang="ts">
import { useHead } from '@unhead/vue'
import { useRoute } from 'vue-router'
import MatchingScene from '@/components/MatchingScene.vue'
import type { MapPoint } from '@/types/map'

// 実験用ダミーデータ
const points: MapPoint[] = [
  { id: '1', lat: 35.689, lng: 139.692, name: 'Alice' },
  { id: '2', lat: 35.694, lng: 139.700, name: 'Bob' },
  { id: '3', lat: 35.683, lng: 139.685, name: 'Carol' },
]

const route = useRoute()
const baseUrl = import.meta.env.VITE_PUBLIC_URL

useHead({
  title: 'InterKnot | ゲームページ',
  meta: [
    { name: 'description', content: 'InterKnotのゲームページです。マッチング状況や参加者の位置情報を確認できます。' },

    { property: 'og:title', content: 'InterKnot | ゲームページ' },
    { property: 'og:description', content: 'InterKnotのゲームページです。マッチング状況や参加者の位置情報を確認できます。' },
    { property: 'og:type', content: 'website' },
    { property: 'og:url', content: `${baseUrl}${route.path}` },
    { property: 'og:image', content: `${baseUrl}/ogp.png` },

    { name: 'twitter:card', content: 'summary_large_image' },
    { name: 'twitter:title', content: 'InterKnot | ゲームページ' },
    { name: 'twitter:description', content: 'InterKnotのゲームページです。マッチング状況や参加者の位置情報を確認できます。' },
    { name: 'twitter:image', content: `${baseUrl}/ogp.png` },
  ],
  link: [
    { rel: 'canonical', href: `${baseUrl}${route.path}` },
  ],
})
</script>

<template>
  <div class="h-svh">
    <MatchingScene
      :current-count="3"
      :max-count="10"
      :points="points"
    />
  </div>
</template>