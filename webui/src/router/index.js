import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import { state } from '../services/state'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/login', name: 'login', component: LoginView },
  ],
})

router.beforeEach((to) => {
  if (!state.userId && to.path !== '/login') {
    return '/login'
  }
  if (state.userId && to.path === '/login') {
    return '/'
  }
  return true
})

export default router
