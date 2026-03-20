<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from '../services/axios'
import { state } from '../services/state'

const router = useRouter()
const username = ref('')
const errorMessage = ref('')
const isLoading = ref(false)

async function doLogin() {
  errorMessage.value = ''

  const name = username.value.trim()
  if (name.length < 3 || name.length > 16) {
    errorMessage.value = 'Display Name deve essere tra 3 e 16 caratteri.'
    return
  }

  isLoading.value = true
  try {
    const response = await axios.post('/session', { name })
    if (response.status === httpStatusCreated && response.data?.identifier) {
      const userId = response.data.identifier
      state.userId = userId
      state.userName = name
      localStorage.setItem('userId', userId)
      localStorage.setItem('userName', name)
      await router.push({ name: 'Home' })
      return
    }

    errorMessage.value = 'Risposta non valida dal server.'
  } catch (err) {
    if (err?.response?.status === 400) {
      errorMessage.value = 'Display Name non valido. Controlla i vincoli richiesti.'
    } else {
      errorMessage.value = 'Errore durante il login. Riprova.'
    }
  } finally {
    isLoading.value = false
  }
}

const httpStatusCreated = 201
</script>

<template>
  <div class="container py-5">
    <div class="row justify-content-center">
      <div class="col-12 col-sm-10 col-md-8 col-lg-5">
        <div class="card shadow-sm border-0">
          <div class="card-body p-4">
            <h1 class="h4 mb-3 text-center">Login</h1>
            <p class="text-muted text-center mb-4">Accedi a WASAText</p>

            <div v-if="errorMessage" class="alert alert-danger" role="alert">
              {{ errorMessage }}
            </div>

            <form @submit.prevent="doLogin">
              <div class="mb-3">
                <label class="form-label" for="displayName">Display Name</label>
                <input
                  id="displayName"
                  v-model.trim="username"
                  type="text"
                  class="form-control"
                  placeholder="Pippo"
                  minlength="3"
                  maxlength="16"
                  required
                />
                <div class="form-text">Min 3, max 16 caratteri.</div>
              </div>

              <button class="btn btn-primary w-100" type="submit" :disabled="isLoading">
                <span v-if="isLoading" class="spinner-border spinner-border-sm me-2" aria-hidden="true"></span>
                Login
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
