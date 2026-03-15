<script setup lang="ts">
</script>

<template>
  <div class="min-h-screen bg-white text-gray-900">
    <header class="border-b border-gray-200 px-6 py-4">
      <nav>
        <RouterLink
          to="/"
          class="font-bold text-lg"
        >
          donguri
        </RouterLink>
      </nav>
    </header>
    <main class="px-6 py-8">
      <slot />
    </main>
  </div>
</template>
