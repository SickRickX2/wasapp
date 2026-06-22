<script setup>
import { Modal } from 'bootstrap'
import { onMounted, onUnmounted, ref, watch } from 'vue'
import axios from '../services/axios'
import { state } from '../services/state'

const emit = defineEmits(['chatCreated'])

const modalRef = ref(null)
const searchQuery = ref('')
const searchResults = ref([])
const errorMessage = ref('')

let debounceTimer = null

async function fetchUsers(query = '') {
  errorMessage.value = ''
  try {
    const suffix = query.trim().length > 0 ? `?q=${encodeURIComponent(query.trim())}` : ''
    const response = await axios.get(`/users${suffix}`)
    const data = response.data
    const users = Array.isArray(data?.users) ? data.users : Array.isArray(data) ? data : []
    // esclude l'utente corrente dai risultati
    searchResults.value = users.filter((user) => user?.userId && user.userId !== state.userId)
  } catch {
    searchResults.value = []
    errorMessage.value = 'Error while searching users.'
  }
}

// carica tutti gli utenti appena il modal diventa visibile
function handleModalShown() {
  fetchUsers(searchQuery.value)
}

onMounted(() => {
  modalRef.value?.addEventListener('shown.bs.modal', handleModalShown)
})

onUnmounted(() => {
  clearTimeout(debounceTimer)
  modalRef.value?.removeEventListener('shown.bs.modal', handleModalShown)
})

watch(searchQuery, (value) => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(async () => {
    await fetchUsers(value)
  }, 300)
})

function resetModalState() {
  searchQuery.value = ''
  searchResults.value = []
  errorMessage.value = ''
}

function cleanupModalArtifacts() {
  document.body.classList.remove('modal-open')
  document.body.style.removeProperty('padding-right')
  document.querySelectorAll('.modal-backdrop').forEach((el) => el.remove())
}

// chiude il modal bootstrap e rimuove manualmente backdrop e classi residue per evitare blocchi allo scroll
function closeModal() {
  if (!modalRef.value) {
    cleanupModalArtifacts()
    return Promise.resolve()
  }

  const instance = Modal.getOrCreateInstance(modalRef.value)
  return new Promise((resolve) => {
    let resolved = false

    const done = () => {
      if (resolved) return
      resolved = true
      cleanupModalArtifacts()
      resolve()
    }

    modalRef.value.addEventListener('hidden.bs.modal', done, { once: true })
    instance.hide()
    setTimeout(done, 350)
  })
}

// post /conversations con recipientId, poi emette chatCreated per navigare alla nuova chat
async function startPrivateChat(recipientId) {
  errorMessage.value = ''
  try {
    const response = await axios.post('/conversations', { recipientId })
    if (response.status === 201) {
      const convId = response.data?.convId
      if (convId) {
        await closeModal()
        resetModalState()
        emit('chatCreated', convId)
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
            <li v-if="searchResults.length === 0" class="list-group-item text-muted">
              No results
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>
