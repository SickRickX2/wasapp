<script setup>
import { RouterLink, RouterView } from 'vue-router'
import { useRouter } from 'vue-router'
import { ref } from 'vue'
import axios from './services/axios'
import { state, updateProfilePictureUrl } from './services/state'
import UserAvatar from './components/UserAvatar.vue'

const router = useRouter()
const profilePictureFileInput = ref(null)
const isUploadingProfilePicture = ref(false)

function normalizeMediaUrl(url) {
  if (!url) return null
  if (/^https?:\/\//i.test(url)) return url
  return `http://localhost:3000${url.startsWith('/') ? '' : '/'}${url}`
}

async function doLogout() {
  try {
    await axios.delete('/session')
  } catch {
    // anche in caso di errore server, forziamo logout lato frontend
  } finally {
    state.userId = null
    state.userName = null
    state.profilePictureUrl = null
    localStorage.removeItem('userId')
    localStorage.removeItem('userName')
    localStorage.removeItem('profilePictureUrl')
    router.push('/login')
  }
}

function triggerProfilePictureInput() {
  profilePictureFileInput.value?.click()
}

async function uploadProfilePicture(event) {
  const file = event.target.files?.[0]
  if (!file) return

  isUploadingProfilePicture.value = true
  try {
    // 1) Upload file to media endpoint (multipart/form-data)
    const formData = new FormData()
    formData.append('file', file)

    const uploadResponse = await axios.post('/media', formData)
    const mediaId = uploadResponse.data?.mediaId

    if (!mediaId) {
      alert('Upload failed: mediaId not returned')
      return
    }

    // 2) Link uploaded media as user profile picture
    const response = await axios.put(`/users/${state.userId}/pfp`, {
      mediaId,
    })

    const pfpUrl = response.data?.pfpUrl
    if (pfpUrl) {
      updateProfilePictureUrl(normalizeMediaUrl(pfpUrl))
      alert('Profile picture updated successfully!')
    } else {
      alert('Profile picture updated, but URL not returned')
    }
  } catch (err) {
    if (err?.response?.status === 413) {
      alert('File too large (max 5MB)')
    } else if (err?.response?.status === 400) {
      alert('Invalid image format or request data')
    } else if (err?.response?.status === 404) {
      alert('Uploaded media not found')
    } else {
      alert('Error uploading profile picture')
    }
  } finally {
    isUploadingProfilePicture.value = false
    event.target.value = ''
  }
}

async function removeProfilePicture() {
  if (!confirm('Remove your profile picture?')) return

  try {
    await axios.delete(`/users/${state.userId}/pfp`)
    updateProfilePictureUrl(null)
    alert('Profile picture removed!')
  } catch (err) {
    alert('Error removing profile picture')
  }
}
</script>

<template>
  <div class="d-flex flex-column vh-100 bg-light overflow-hidden">
    <template v-if="state.userId">
      <nav class="navbar navbar-expand-lg navbar-dark bg-dark px-3">
        <span class="navbar-brand mb-0 h1">WASAText</span>
        <div class="ms-auto d-flex align-items-center gap-3">
          <div class="d-flex align-items-center gap-2">
            <UserAvatar :name="state.userId" :displayName="state.userName" :realImageUrl="state.profilePictureUrl" :size="32" />
            <small class="text-white-50">{{ state.userName || 'User' }}</small>
          </div>
          <button 
            class="btn btn-outline-light btn-sm" 
            type="button" 
            data-bs-toggle="modal" 
            data-bs-target="#profileSettingsModal"
            title="Profile settings"
          >
            <span class="material-symbols-outlined" style="font-size: 18px;">settings</span>
          </button>
          <button class="btn btn-outline-light btn-sm" @click="doLogout">Logout</button>
        </div>
      </nav>

      <div class="container-fluid flex-grow-1 overflow-hidden" style="min-height: 0;">
        <div class="row h-100 gx-0" style="min-height: 0;">
          <aside class="col-12 col-md-3 col-lg-2 bg-white border-end h-100 p-3">
            <div class="list-group list-group-flush">
              <RouterLink to="/" class="list-group-item list-group-item-action">Home</RouterLink>
              <button type="button" class="list-group-item list-group-item-action" data-bs-toggle="modal" data-bs-target="#newChatModal">
                <span class="material-symbols-outlined align-middle me-2" style="font-size: 20px;">chat</span> Nuova Chat
              </button>
              <button type="button" class="list-group-item list-group-item-action" data-bs-toggle="modal" data-bs-target="#newGroupModal">
                <span class="material-symbols-outlined align-middle me-2" style="font-size: 20px;">group_add</span> Nuovo Gruppo
              </button>
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

    <!-- Profile Settings Modal -->
    <div class="modal fade" id="profileSettingsModal" tabindex="-1" aria-labelledby="profileSettingsLabel" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h1 class="modal-title fw-bold" id="profileSettingsLabel" style="font-size: 1.25rem;">Settings</h1>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body text-center" style="font-size: 0.95rem;">
            <div class="d-block mx-auto mb-3" style="width: fit-content;">
              <UserAvatar :name="state.userId" :displayName="state.userName" :realImageUrl="state.profilePictureUrl" :size="128" />
            </div>
            <div class="d-flex justify-content-center gap-2">
              <button 
                class="btn btn-primary btn-sm" 
                type="button" 
                @click="triggerProfilePictureInput"
                :disabled="isUploadingProfilePicture"
              >
                <span v-if="isUploadingProfilePicture" class="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
                {{ isUploadingProfilePicture ? 'Caricamento...' : 'Upload Photo' }}
              </button>
              <button 
                v-if="state.profilePictureUrl"
                class="btn btn-outline-danger btn-sm" 
                type="button" 
                @click="removeProfilePicture"
                :disabled="isUploadingProfilePicture"
              >
                Rimuovi Foto
              </button>
            </div>
            <input
              ref="profilePictureFileInput"
              type="file"
              class="d-none"
              accept="image/*"
              @change="uploadProfilePicture"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
.material-symbols-outlined {
  font-family: 'Material Symbols Outlined';
  font-weight: normal;
  font-style: normal;
  font-size: 24px;
  line-height: 1;
  letter-spacing: normal;
  text-transform: none;
  display: inline-block;
  white-space: nowrap;
  word-wrap: normal;
  direction: ltr;
  -webkit-font-feature-settings: 'liga';
  -webkit-font-smoothing: antialiased;
}
</style>
