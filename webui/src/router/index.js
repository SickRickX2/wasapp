import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import { state } from '../services/state'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'Home', component: HomeView },
    { path: '/login', name: 'Login', component: LoginView },
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
