<script setup>
import { RouterLink, RouterView } from 'vue-router'
import { useRouter } from 'vue-router'
import axios from './services/axios'
import { state } from './services/state'

const router = useRouter()

async function doLogout() {
  try {
    await axios.delete('/session')
  } catch {
    // anche in caso di errore server, forziamo logout lato frontend
  } finally {
    state.userId = null
    state.userName = null
    localStorage.removeItem('userId')
    localStorage.removeItem('userName')
    router.push('/login')
  }
}
</script>

<template>
  <div class="d-flex flex-column vh-100 bg-light overflow-hidden">
    <template v-if="state.userId">
      <nav class="navbar navbar-expand-lg navbar-dark bg-dark px-3">
        <span class="navbar-brand mb-0 h1">WASAText</span>
        <div class="ms-auto d-flex align-items-center gap-3 text-white">
          <small class="text-white-50">{{ state.userName || 'User' }}</small>
          <button class="btn btn-outline-light btn-sm" @click="doLogout">Logout</button>
        </div>
      </nav>

      <div class="container-fluid flex-grow-1 overflow-hidden" style="min-height: 0;">
        <div class="row h-100 gx-0" style="min-height: 0;">
          <aside class="col-12 col-md-3 col-lg-2 bg-white border-end h-100 p-3">
            <div class="list-group list-group-flush">
              <RouterLink to="/" class="list-group-item list-group-item-action">Home</RouterLink>
            </div>
          </aside>

          <main class="col-12 col-md-9 col-lg-10 h-100 p-0 overflow-hidden" style="min-height: 0;">
            <RouterView />
          </main>
        </div>
      </div>
    </template>

    <template v-else>
      <RouterView />
    </template>
  </div>
</template>
