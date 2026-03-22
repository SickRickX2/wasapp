<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import axios from '../services/axios'
import { state } from '../services/state'
import NewChatModal from '../components/NewChatModal.vue'
import UserAvatar from '../components/UserAvatar.vue'

const conversations = ref([])
const currentMessages = ref([])
const selectedConversationId = ref(null)
const newMessageText = ref('')
const userNameCache = ref({})
const fileInput = ref(null)
const isUploadingMedia = ref(false)
const previewImageUrl = ref('')
const selectedMediaFile = ref(null)
const selectedMediaPreviewUrl = ref('')
const messagesContainer = ref(null)
let pollingInterval = null

const selectedConversation = computed(() => {
  return conversations.value.find((c) => c.convId === selectedConversationId.value) ?? null
})

const sortedMessages = computed(() => {
  return currentMessages.value.slice().reverse()
})

const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

watch(
  currentMessages,
  () => {
    scrollToBottom()
  },
  { deep: true }
)

function getUsernameFromId(userId) {
  if (!userId) return ''
  if (userNameCache.value[userId]) return userNameCache.value[userId]

  const selected = selectedConversation.value
  const selectedParticipants = Array.isArray(selected?.participants) ? selected.participants : []
  const selectedNames = Array.isArray(selected?.participantNames) ? selected.participantNames : []
  const selectedIndex = selectedParticipants.findIndex((participantId) => participantId === userId)
  if (selectedIndex >= 0 && selectedNames[selectedIndex]) {
    return selectedNames[selectedIndex]
  }

  for (const conv of conversations.value) {
    const participants = Array.isArray(conv?.participants) ? conv.participants : []
    const names = Array.isArray(conv?.participantNames) ? conv.participantNames : []
    const index = participants.findIndex((participantId) => participantId === userId)
    if (index >= 0 && names[index]) {
      return names[index]
    }
  }

  return ''
}

async function resolveUsernameById(userId) {
  if (!userId) return ''
  if (userId === state.userId) {
    return state.userName || 'Me'
  }
  if (userNameCache.value[userId]) {
    return userNameCache.value[userId]
  }

  try {
    const response = await axios.get(`/users/${userId}`)
    const userName = response.data?.userName || ''
    if (userName) {
      userNameCache.value[userId] = userName
    }
    return userName
  } catch {
    return ''
  }
}

async function enrichConversationNames() {
  const idsToResolve = new Set()

  for (const conv of conversations.value) {
    const participants = Array.isArray(conv?.participants) ? conv.participants : []
    const names = Array.isArray(conv?.participantNames) ? conv.participantNames : []

    participants.forEach((participantId, index) => {
      const participantName = names[index]
      if (participantName) {
        userNameCache.value[participantId] = participantName
      } else if (participantId && participantId !== state.userId) {
        idsToResolve.add(participantId)
      }
    })
  }

  if (idsToResolve.size > 0) {
    await Promise.all(Array.from(idsToResolve).map((id) => resolveUsernameById(id)))
  }
}

function getChatTitle(conv) {
  if (!conv) return ''
  if (conv.type === 'group') {
    return conv.groupName || 'Group'
  }
  if (conv.type === 'private') {
    const participants = Array.isArray(conv.participants) ? conv.participants : []
    const participantNames = Array.isArray(conv.participantNames) ? conv.participantNames : []
    const otherIndex = participants.findIndex((participantId) => participantId !== state.userId)
    if (otherIndex >= 0 && participantNames[otherIndex]) {
      return participantNames[otherIndex]
    }
    const otherUserId = participants[otherIndex]
    const otherUserName = getUsernameFromId(otherUserId)
    if (otherUserName) {
      return otherUserName
    }
    return otherUserId || 'Private chat'
  }
  return 'Chat'
}

function getSenderLabel(message) {
  if (!message?.sender) return 'User'
  if (message.sender === state.userId) return 'Me'
  const senderName = getUsernameFromId(message.sender)
  if (senderName) return senderName
  return message.sender
}

function getMediaUrl(rawUrl) {
  if (!rawUrl) return ''
  if (rawUrl.startsWith('http://') || rawUrl.startsWith('https://')) {
    return rawUrl
  }

  const baseUrl = (axios.defaults.baseURL || '').replace(/\/$/, '')
  const path = rawUrl.startsWith('/') ? rawUrl : `/${rawUrl}`
  return `${baseUrl}${path}`
}

function openImagePreview(rawUrl) {
  const resolvedUrl = getMediaUrl(rawUrl)
  if (!resolvedUrl) return
  previewImageUrl.value = resolvedUrl
}

function closeImagePreview() {
  previewImageUrl.value = ''
}

async function loadConversations() {
  try {
    if (!state.userId) return
    const response = await axios.get(`/users/${state.userId}/conversations`)
    const data = response.data
    conversations.value = Array.isArray(data?.conversations) ? data.conversations : []
    await enrichConversationNames()
  } catch (err) {
    console.error('Failed to load conversations', err)
  }
}

async function loadMessages(conversationId) {
  try {
    const response = await axios.get(`/conversations/${conversationId}/messages`)
    const data = response.data
    if (Array.isArray(data?.messages)) {
      currentMessages.value = data.messages
      const senderIds = new Set(data.messages.map((m) => m?.sender).filter(Boolean))
      await Promise.all(Array.from(senderIds).map((id) => resolveUsernameById(id)))
      await scrollToBottom()
      return
    }
    currentMessages.value = Array.isArray(data) ? data : []
    const senderIds = new Set(currentMessages.value.map((m) => m?.sender).filter(Boolean))
    await Promise.all(Array.from(senderIds).map((id) => resolveUsernameById(id)))
    await scrollToBottom()
  } catch (err) {
    console.error('Failed to load messages', err)
  }
}

async function syncData() {
  await loadConversations()
  if (selectedConversationId.value) {
    await loadMessages(selectedConversationId.value)
  }
}

async function selectConversation(id) {
  clearSelectedMedia()
  selectedConversationId.value = id
  await loadMessages(id)
}

function triggerFileInput() {
  fileInput.value?.click()
}

async function uploadAndSendMedia(event) {
  const file = event.target.files?.[0]
  if (!file) return

  if (!selectedConversationId.value) {
    alert('Select a chat before sending media')
    event.target.value = ''
    return
  }

  if (selectedMediaPreviewUrl.value) {
    URL.revokeObjectURL(selectedMediaPreviewUrl.value)
  }

  selectedMediaFile.value = file
  selectedMediaPreviewUrl.value = URL.createObjectURL(file)
  event.target.value = ''
}

function clearSelectedMedia() {
  if (selectedMediaPreviewUrl.value) {
    URL.revokeObjectURL(selectedMediaPreviewUrl.value)
  }
  selectedMediaPreviewUrl.value = ''
  selectedMediaFile.value = null
}

async function sendMessage() {
  if (!selectedConversationId.value) return
  const text = newMessageText.value.trim()
  const hasMedia = !!selectedMediaFile.value
  if (text.length === 0 && !hasMedia) return

  isUploadingMedia.value = true
  try {
    let mediaPayload = null

    if (selectedMediaFile.value) {
      const formData = new FormData()
      formData.append('file', selectedMediaFile.value)

      const uploadResponse = await axios.post('/media', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })

      const mediaUrl = uploadResponse.data?.url
      if (!mediaUrl) {
        alert('Error: media URL not received from server')
        return
      }

      mediaPayload = {
        url: mediaUrl,
        filename: selectedMediaFile.value.name,
        mimeType: selectedMediaFile.value.type,
        size: selectedMediaFile.value.size,
      }
    }

    const payload = {}
    if (text.length > 0) {
      payload.text = text
    }
    if (mediaPayload) {
      payload.media = mediaPayload
    }

    await axios.post(`/conversations/${selectedConversationId.value}/messages`, payload)
    newMessageText.value = ''
    clearSelectedMedia()
    await loadMessages(selectedConversationId.value)
  } catch (err) {
    if (err?.response?.status === 413) {
      alert('File too large (max 5MB)')
    } else if (err?.response?.status === 400) {
      alert('Invalid message data')
    } else {
      alert('Error while sending message')
    }
  } finally {
    isUploadingMedia.value = false
  }
}

async function onChatCreated(convId) {
  await loadConversations()
  selectedConversationId.value = convId
  await loadMessages(convId)
}

onMounted(async () => {
  await loadConversations()
  pollingInterval = setInterval(syncData, 3000)
})

onUnmounted(() => {
  if (pollingInterval) clearInterval(pollingInterval)
})
</script>

<template>
  <div class="d-flex h-100 overflow-hidden bg-white" style="min-height: 0;">
    <aside class="d-flex flex-column h-100 border-end chat-sidebar">
      <div class="p-3 border-bottom d-flex align-items-center justify-content-between gap-2">
        <h2 class="h5 mb-0">Chat</h2>
        <button
          class="btn btn-sm btn-primary"
          type="button"
          data-bs-toggle="modal"
          data-bs-target="#newChatModal"
        >
          New Chat
        </button>
      </div>

      <div class="list-group list-group-flush overflow-y-auto flex-grow-1">
        <button
          v-for="conversation in conversations"
          :key="conversation.convId"
          type="button"
          class="list-group-item list-group-item-action d-flex align-items-center gap-3"
          :class="{ active: selectedConversationId === conversation.convId }"
          @click="selectConversation(conversation.convId)"
        >
          <UserAvatar 
            :name="conversation.type === 'group' ? conversation.groupName || 'Group' : conversation.participants?.find(pid => pid !== state.userId) || 'user'" 
            :displayName="conversation.type === 'group' ? conversation.groupName || 'Group' : getUsernameFromId(conversation.participants?.find(pid => pid !== state.userId))"
            :size="40"
            :realImageUrl="null"
            class="flex-shrink-0"
          />
          <div class="flex-grow-1 min-w-0">
            <div class="fw-semibold text-truncate">{{ getChatTitle(conversation) }}</div>
            <small class="text-muted" :class="{ 'text-white-50': selectedConversationId === conversation.convId }">
              {{ conversation.type }}
            </small>
          </div>
        </button>
      </div>
    </aside>

    <section class="d-flex flex-column h-100 min-w-0 flex-grow-1" style="min-height: 0;">
      <template v-if="!selectedConversationId">
        <div class="d-flex flex-grow-1 align-items-center justify-content-center text-muted">
          Select a chat to start
        </div>
      </template>

      <template v-else>
        <header class="border-bottom p-3 bg-white flex-shrink-0 d-flex align-items-center gap-3">
          <UserAvatar 
            :name="selectedConversation?.type === 'group' ? selectedConversation?.groupName || 'Group' : selectedConversation?.participants?.find(pid => pid !== state.userId) || 'user'" 
            :displayName="selectedConversation?.type === 'group' ? selectedConversation?.groupName || 'Group' : getUsernameFromId(selectedConversation?.participants?.find(pid => pid !== state.userId))"
            :size="40"
            :realImageUrl="null"
          />
          <h3 class="h6 mb-0">{{ getChatTitle(selectedConversation) }}</h3>
        </header>

        <div ref="messagesContainer" class="flex-grow-1 overflow-y-auto p-3" style="min-height: 0;">
          <div
            v-for="(message, index) in sortedMessages"
            :key="message.messageId || index"
            class="d-flex mb-2"
            :class="message.sender === state.userId ? 'justify-content-end' : 'justify-content-start'"
          >
            <div
              class="px-3 py-2 rounded-3 message-bubble"
              :class="message.sender === state.userId ? 'bg-primary text-white' : 'bg-white border'"
            >
              <div class="small mb-1 opacity-75">{{ getSenderLabel(message) }}</div>
              <div v-if="message.text">{{ message.text }}</div>
              <img
                v-if="message.media?.url"
                :src="getMediaUrl(message.media.url)"
                alt="media"
                class="img-fluid rounded mt-2 media-thumb"
                style="max-height: 220px"
                role="button"
                @click="openImagePreview(message.media.url)"
              />
            </div>
          </div>
        </div>

        <div class="p-3 bg-light border-top mt-auto flex-shrink-0">
          <div class="input-group">
            <button class="btn btn-outline-secondary d-flex align-items-center" type="button" @click="triggerFileInput">
              <span class="material-symbols-outlined">add_photo_alternate</span>
            </button>
            <input type="text" class="form-control" placeholder="Type a message..." v-model="newMessageText" @keyup.enter="sendMessage">
            <button class="btn btn-primary" type="button" @click="sendMessage">Send</button>
          </div>
          <input
            ref="fileInput"
            type="file"
            class="d-none"
            accept="image/*"
            @change="uploadAndSendMedia"
          />

          <div v-if="selectedMediaFile" class="mt-2 d-inline-flex align-items-center gap-2 border rounded p-2 bg-white">
            <img
              :src="selectedMediaPreviewUrl"
              alt="Attachment preview"
              class="rounded"
              style="width: 44px; height: 44px; object-fit: cover"
            />
            <small class="text-muted text-truncate" style="max-width: 220px">{{ selectedMediaFile.name }}</small>
            <button type="button" class="btn btn-sm btn-outline-danger" @click="clearSelectedMedia">✕</button>
          </div>
        </div>
      </template>
    </section>

    <NewChatModal @chatCreated="onChatCreated" />

    <div v-if="previewImageUrl" class="image-preview-overlay" @click="closeImagePreview">
      <button class="btn btn-light image-preview-close" type="button" @click.stop="closeImagePreview">✕</button>
      <img :src="previewImageUrl" alt="Image preview" class="image-preview-full" @click.stop />
    </div>
  </div>
</template>

<style scoped>
.chat-sidebar {
  width: 32%;
  min-width: 260px;
  max-width: 420px;
}

.message-bubble {
  max-width: 75%;
  word-break: break-word;
}

.material-symbols-outlined {
  font-variation-settings: 'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24;
  font-size: 24px;
  vertical-align: middle;
}

.media-thumb {
  cursor: zoom-in;
}

.image-preview-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.9);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  padding: 1rem;
}

.image-preview-full {
  max-width: 95vw;
  max-height: 92vh;
  object-fit: contain;
  border-radius: 0.5rem;
  box-shadow: 0 0 30px rgba(0, 0, 0, 0.4);
}

.image-preview-close {
  position: absolute;
  top: 1rem;
  right: 1rem;
  z-index: 2001;
}
</style>
