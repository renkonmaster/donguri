<script setup lang="ts">
import { useHead } from '@unhead/vue'
import { useRoute } from 'vue-router'
import DefaultLayout from '@/layouts/DefaultLayout.vue'

const route = useRoute()
const baseUrl = import.meta.env.VITE_PUBLIC_URL

useHead({
  title: '404 Not Found | InterKnot',
  meta: [
    { name: 'description', content: 'ページが見つかりませんでした。' },

    { property: 'og:title', content: '404 Not Found | InterKnot' },
    { property: 'og:description', content: 'ページが見つかりませんでした。' },
    { property: 'og:type', content: 'website' },
    { property: 'og:url', content: `${baseUrl}${route.path}` },
    { property: 'og:image', content: `${baseUrl}/ogp.png` },

    { name: 'twitter:card', content: 'summary_large_image' },
    { name: 'twitter:title', content: '404 Not Found | InterKnot' },
    { name: 'twitter:description', content: 'ページが見つかりませんでした。' },
    { name: 'twitter:image', content: `${baseUrl}/ogp.png` },
  ],
  link: [
    { rel: 'canonical', href: `${baseUrl}${route.path}` },
  ],
})
</script>

<template>
  <DefaultLayout>
    <div class="flex flex-col items-center gap-4 py-16">
      <p class="text-6xl font-bold text-gray-300">
        404
      </p>
      <p class="text-gray-600">
        ページが見つかりませんでした。
      </p>
      <RouterLink
        to="/"
        class="text-blue-600 underline"
      >
        トップへ戻る
      </RouterLink>
    </div>
  </DefaultLayout>
</template>