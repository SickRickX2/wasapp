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

  const name = username.value
  if (name.length < 3 || name.length > 16) {
    errorMessage.value = 'User Name must be between 3 and 16 characters.'
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

    errorMessage.value = 'Invalid response from server.'
  } catch (err) {
    if (err?.response?.status === 400) {
      errorMessage.value = 'Invalid User Name. Check required constraints.'
    } else {
      errorMessage.value = 'Login failed. Please try again.'
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
            <p class="text-muted text-center mb-4">Sign in to WASAText</p>

            <div v-if="errorMessage" class="alert alert-danger" role="alert">
              {{ errorMessage }}
            </div>

            <form @submit.prevent="doLogin">
              <div class="mb-3">
                <label class="form-label" for="displayName">Username</label>
                <input
                  id="displayName"
                  v-model="username"
                  type="text"
                  class="form-control"
                  minlength="3"
                  maxlength="16"
                  required
                  oninvalid="this.setCustomValidity('Please fill out this field.')"
                  oninput="this.setCustomValidity('')"
                />
                <div class="form-text">Min 3, max 16 characters.</div>
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
