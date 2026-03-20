<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { state } from '../services/state'

const router = useRouter()
const userId = ref('')

function doLogin() {
  const id = userId.value.trim()
  if (!id) return
  state.userId = id
  router.push('/')
}
</script>

<template>
  <div class="container py-5">
    <div class="row justify-content-center">
      <div class="col-12 col-md-6 col-lg-4">
        <div class="card shadow-sm">
          <div class="card-body">
            <h1 class="h4 mb-3">Login</h1>
            <p class="text-muted small">Inserisci lo userId (Bearer token) per continuare.</p>

            <div class="mb-3">
              <label class="form-label">User ID</label>
              <input
                v-model="userId"
                type="text"
                class="form-control"
                placeholder="usr_a1B2c3"
                @keyup.enter="doLogin"
              />
            </div>

            <button class="btn btn-primary w-100" @click="doLogin">Entra</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
