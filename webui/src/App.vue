<script setup>
import { computed } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { state } from './services/state'

const isLoggedIn = computed(() => state.userId !== null)

function logout() {
  state.userId = null
  state.userName = null
  localStorage.removeItem('userId')
  localStorage.removeItem('userName')
}
</script>

<template>
  <div class="d-flex flex-column vh-100 bg-light">
    <template v-if="isLoggedIn">
      <nav class="navbar navbar-expand-lg navbar-dark bg-dark px-3">
        <span class="navbar-brand mb-0 h1">WASAText</span>
        <div class="ms-auto d-flex align-items-center gap-3 text-white">
          <small class="text-white-50">{{ state.userName || 'Utente' }}</small>
          <button class="btn btn-outline-light btn-sm" @click="logout">Logout</button>
        </div>
      </nav>

      <div class="container-fluid flex-grow-1 min-vh-0">
        <div class="row h-100 gx-0">
          <aside class="col-12 col-md-3 col-lg-2 bg-white border-end h-100 p-3">
            <div class="list-group list-group-flush">
              <RouterLink to="/" class="list-group-item list-group-item-action">Home</RouterLink>
            </div>
          </aside>

          <main class="col-12 col-md-9 col-lg-10 h-100 p-0">
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
