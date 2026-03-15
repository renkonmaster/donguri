import { useRoute } from 'vue-router';
import { useHead } from '@unhead/vue';

export function useOgpHead(title: string, description: string) {
  const route = useRoute();
  const baseUrl = import.meta.env.VITE_PUBLIC_URL;

  useHead({
    title,
    meta: [
      { name: 'description', content: description },

      { property: 'og:title', content: title },
      { property: 'og:description', content: description },
      { property: 'og:type', content: 'website' },
      { property: 'og:url', content: `${baseUrl}${route.path}` },
      { property: 'og:image', content: `${baseUrl}/ogp.png` },

      { name: 'twitter:card', content: 'summary_large_image' },
      { name: 'twitter:title', content: title },
      { name: 'twitter:description', content: description },
      { name: 'twitter:image', content: `${baseUrl}/ogp.png` },
    ],
    link: [
      { rel: 'canonical', href: `${baseUrl}${route.path}` },
    ],
  });
}
