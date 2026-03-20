<script setup>
import { Modal } from 'bootstrap'
import { ref, watch } from 'vue'
import axios from '../services/axios'

const emit = defineEmits(['chatCreated'])

const modalRef = ref(null)
const searchQuery = ref('')
const searchResults = ref([])
const errorMessage = ref('')

let debounceTimer = null

watch(searchQuery, (value) => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(async () => {
    const q = value.trim()
    errorMessage.value = ''

    if (q.length === 0) {
      searchResults.value = []
      return
    }

    try {
      const response = await axios.get(`/users?q=${encodeURIComponent(q)}`)
      const data = response.data
      searchResults.value = Array.isArray(data?.users) ? data.users : Array.isArray(data) ? data : []
    } catch {
      searchResults.value = []
      errorMessage.value = 'Error while searching users.'
    }
  }, 300)
})

function resetModalState() {
  searchQuery.value = ''
  searchResults.value = []
  errorMessage.value = ''
}

function closeModal() {
  if (!modalRef.value) return
  const instance = Modal.getOrCreateInstance(modalRef.value)
  instance.hide()
}

async function startPrivateChat(recipientId) {
  errorMessage.value = ''
  try {
    const response = await axios.post('/conversations', { recipientId })
    if (response.status === 201) {
      const convId = response.data?.convId
      if (convId) {
        emit('chatCreated', convId)
        closeModal()
        resetModalState()
      }
    }
  } catch {
    errorMessage.value = 'Unable to create private chat.'
  }
}
</script>

<template>
  <div
    id="newChatModal"
    ref="modalRef"
    class="modal fade"
    tabindex="-1"
    aria-labelledby="newChatModalLabel"
    aria-hidden="true"
  >
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h5 id="newChatModalLabel" class="modal-title">New chat</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close" @click="resetModalState" />
        </div>

        <div class="modal-body">
          <input v-model="searchQuery" type="text" class="form-control mb-3" placeholder="Search user..." />

          <div v-if="errorMessage" class="alert alert-danger py-2" role="alert">
            {{ errorMessage }}
          </div>

          <ul class="list-group">
            <li
              v-for="user in searchResults"
              :key="user.userId"
              class="list-group-item list-group-item-action"
              role="button"
              @click="startPrivateChat(user.userId)"
            >
              {{ user.userName }}
            </li>
            <li v-if="searchQuery.trim().length > 0 && searchResults.length === 0" class="list-group-item text-muted">
              No results
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>
