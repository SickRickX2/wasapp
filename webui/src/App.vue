<script setup>
import { computed } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { state } from './services/state'

const isLoggedIn = computed(() => state.userId !== null)

function logout() {
  state.userId = null
}
</script>

<template>
  <div class="min-vh-100 bg-light">
    <template v-if="isLoggedIn">
      <nav class="navbar navbar-expand-lg navbar-dark bg-dark px-3">
        <span class="navbar-brand mb-0 h1">WASAText</span>
        <div class="ms-auto d-flex align-items-center gap-3 text-white">
          <small class="text-white-50">{{ state.userId }}</small>
          <button class="btn btn-outline-light btn-sm" @click="logout">Logout</button>
        </div>
      </nav>

      <div class="container-fluid">
        <div class="row">
          <aside class="col-12 col-md-3 col-lg-2 bg-white border-end min-vh-100 p-3">
            <div class="list-group list-group-flush">
              <RouterLink to="/" class="list-group-item list-group-item-action">Home</RouterLink>
            </div>
          </aside>

          <main class="col-12 col-md-9 col-lg-10 p-4">
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
