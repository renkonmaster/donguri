<script setup lang="ts">
import { useOgpHead } from '@/composables/useOgpHead';
import DefaultLayout from '@/layouts/DefaultLayout.vue';

useOgpHead(
  '404 Not Found | InterKnot',
  'ページが見つかりませんでした。',
);
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
