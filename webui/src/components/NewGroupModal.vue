<script setup>
import { Modal } from 'bootstrap'
import { onUnmounted, ref, watch } from 'vue'
import axios from '../services/axios'

const emit = defineEmits(['groupCreated'])

const modalRef = ref(null)
const groupName = ref('')
const searchQuery = ref('')
const searchResults = ref([])
const selectedUsers = ref([])
const groupPhotoInputRef = ref(null)
const selectedGroupPhotoFile = ref(null)
const errorMessage = ref('')
const isSubmitting = ref(false)

let debounceTimer = null

onUnmounted(() => {
  clearTimeout(debounceTimer)
})

watch(searchQuery, (value) => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(async () => {
    const q = value.trim()
    errorMessage.value = ''

    if (!q) {
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

function isSelected(userId) {
  return selectedUsers.value.some((u) => u.userId === userId)
}

// aggiunge o rimuove l'utente dalla selezione in base alla sua presenza corrente
function toggleUser(user) {
  if (!user?.userId) return
  const index = selectedUsers.value.findIndex((u) => u.userId === user.userId)
  if (index >= 0) {
    selectedUsers.value.splice(index, 1)
    return
  }
  selectedUsers.value.push(user)
}

function removeSelectedUser(userId) {
  selectedUsers.value = selectedUsers.value.filter((u) => u.userId !== userId)
}

function triggerGroupPhotoInput() {
  groupPhotoInputRef.value?.click()
}

function onGroupPhotoSelected(event) {
  const file = event.target.files?.[0] || null
  selectedGroupPhotoFile.value = file
  event.target.value = ''
}

function clearGroupPhoto() {
  selectedGroupPhotoFile.value = null
}

function resetModalState() {
  groupName.value = ''
  searchQuery.value = ''
  searchResults.value = []
  selectedUsers.value = []
  selectedGroupPhotoFile.value = null
  errorMessage.value = ''
  isSubmitting.value = false
}

function cleanupModalArtifacts() {
  document.body.classList.remove('modal-open')
  document.body.style.removeProperty('padding-right')
  document.querySelectorAll('.modal-backdrop').forEach((el) => el.remove())
}

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

// crea prima il gruppo con post /groups, poi se c'è una foto la carica separatamente con una put
async function submitGroup() {
  const trimmedGroupName = groupName.value.trim()
  if (!trimmedGroupName) {
    errorMessage.value = 'Group name is required.'
    return
  }

  isSubmitting.value = true
  errorMessage.value = ''
  try {
    const participants = selectedUsers.value.map((u) => u.userId)
    const response = await axios.post('/groups', {
      groupName: trimmedGroupName,
      participants,
    })

    if (response.status === 201) {
      const convId = response.data?.convId
      if (convId && selectedGroupPhotoFile.value) {
        const groupPhotoData = new FormData()
        groupPhotoData.append('file', selectedGroupPhotoFile.value)
        await axios.put(`/conversations/${convId}/group_photo`, groupPhotoData)
      }

      await closeModal()
      resetModalState()

      if (convId) {
        emit('groupCreated', convId)
      }
      return
    }

    errorMessage.value = 'Unable to create group.'
  } catch {
    errorMessage.value = 'Unable to create group.'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div
    id="newGroupModal"
    ref="modalRef"
    class="modal fade"
    tabindex="-1"
    aria-labelledby="newGroupModalLabel"
    aria-hidden="true"
  >
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h5 id="newGroupModalLabel" class="modal-title">New Group</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close" @click="resetModalState" />
        </div>

        <div class="modal-body">
          <label class="form-label">Group Name</label>
          <input v-model="groupName" type="text" class="form-control mb-3" placeholder="Enter group name" />

          <div class="mb-3">
            <label class="form-label d-block">Group Image (optional)</label>
            <div class="d-flex align-items-center gap-2">
              <button type="button" class="btn btn-outline-secondary btn-sm" @click="triggerGroupPhotoInput">
                Upload image
              </button>
              <small v-if="selectedGroupPhotoFile" class="text-muted text-truncate" style="max-width: 220px;">
                {{ selectedGroupPhotoFile.name }}
              </small>
              <button v-if="selectedGroupPhotoFile" type="button" class="btn btn-sm btn-outline-danger" @click="clearGroupPhoto">
                ✕
              </button>
            </div>
            <input
              ref="groupPhotoInputRef"
              type="file"
              class="d-none"
              accept="image/*"
              @change="onGroupPhotoSelected"
            />
          </div>

          <div class="mb-2 d-flex flex-wrap gap-2" v-if="selectedUsers.length > 0">
            <span
              v-for="user in selectedUsers"
              :key="user.userId"
              class="badge rounded-pill text-bg-primary"
              role="button"
              @click="removeSelectedUser(user.userId)"
              :title="`Remove ${user.userName}`"
            >
              {{ user.userName }} ✕
            </span>
          </div>

          <label class="form-label">Search users</label>
          <input v-model="searchQuery" type="text" class="form-control mb-3" placeholder="Search user..." />

          <ul class="list-group">
            <li
              v-for="user in searchResults"
              :key="user.userId"
              class="list-group-item list-group-item-action d-flex justify-content-between align-items-center"
              :class="{ active: isSelected(user.userId) }"
              role="button"
              @click="toggleUser(user)"
            >
              <span>{{ user.userName }}</span>
              <span class="small" v-if="isSelected(user.userId)">Selected</span>
            </li>
            <li v-if="searchQuery.trim().length > 0 && searchResults.length === 0" class="list-group-item text-muted">
              No results
            </li>
          </ul>

          <div v-if="errorMessage" class="alert alert-danger py-2 mt-3 mb-0" role="alert">
            {{ errorMessage }}
          </div>
        </div>

        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" data-bs-dismiss="modal" @click="resetModalState">Cancel</button>
          <button type="button" class="btn btn-primary" :disabled="isSubmitting" @click="submitGroup">
            {{ isSubmitting ? 'Creating...' : 'Create Group' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
