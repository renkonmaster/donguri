import type { RouteRecordRaw } from 'vue-router';

export const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'index',
    component: () => import('@/pages/IndexPage.vue'),
  },
  {
    path: '/game',
    name: 'game',
    component: () => import('@/pages/GamePage.vue'),
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
];
