<script setup>
import { computed, onMounted, ref } from 'vue'
import axios from '../services/axios'
import { state } from '../services/state'
import NewChatModal from '../components/NewChatModal.vue'

const conversations = ref([])
const currentMessages = ref([])
const selectedConversationId = ref(null)
const newMessageText = ref('')
const userNameCache = ref({})

const selectedConversation = computed(() => {
  return conversations.value.find((c) => c.convId === selectedConversationId.value) ?? null
})

const sortedMessages = computed(() => {
  return currentMessages.value.slice().reverse()
})

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
    return otherUserId || 'Chat privata'
  }
  return 'Chat'
}

function getSenderLabel(message) {
  if (!message?.sender) return 'Utente'
  if (message.sender === state.userId) return 'Me'
  const senderName = getUsernameFromId(message.sender)
  if (senderName) return senderName
  return message.sender
}

async function loadConversations() {
  if (!state.userId) return
  const response = await axios.get(`/users/${state.userId}/conversations`)
  const data = response.data
  conversations.value = Array.isArray(data?.conversations) ? data.conversations : []
  await enrichConversationNames()
}

async function loadMessages(conversationId) {
  const response = await axios.get(`/conversations/${conversationId}/messages`)
  const data = response.data
  if (Array.isArray(data?.messages)) {
    currentMessages.value = data.messages
    const senderIds = new Set(data.messages.map((m) => m?.sender).filter(Boolean))
    await Promise.all(Array.from(senderIds).map((id) => resolveUsernameById(id)))
    return
  }
  currentMessages.value = Array.isArray(data) ? data : []
  const senderIds = new Set(currentMessages.value.map((m) => m?.sender).filter(Boolean))
  await Promise.all(Array.from(senderIds).map((id) => resolveUsernameById(id)))
}

async function selectConversation(id) {
  selectedConversationId.value = id
  await loadMessages(id)
}

async function sendMessage() {
  if (!selectedConversationId.value) return
  if (newMessageText.value.trim().length === 0) return

  await axios.post(`/conversations/${selectedConversationId.value}/messages`, {
    text: newMessageText.value,
  })

  newMessageText.value = ''
  await loadMessages(selectedConversationId.value)
}

async function onChatCreated(convId) {
  await loadConversations()
  selectedConversationId.value = convId
  await loadMessages(convId)
}

onMounted(async () => {
  await loadConversations()
})
</script>

<template>
  <div class="d-flex h-100 overflow-hidden bg-white">
    <aside class="d-flex flex-column h-100 border-end chat-sidebar">
      <div class="p-3 border-bottom d-flex align-items-center justify-content-between gap-2">
        <h2 class="h5 mb-0">Chat</h2>
        <button
          class="btn btn-sm btn-primary"
          type="button"
          data-bs-toggle="modal"
          data-bs-target="#newChatModal"
        >
          Nuova Chat
        </button>
      </div>

      <div class="list-group list-group-flush overflow-y-auto flex-grow-1">
        <button
          v-for="conversation in conversations"
          :key="conversation.convId"
          type="button"
          class="list-group-item list-group-item-action"
          :class="{ active: selectedConversationId === conversation.convId }"
          @click="selectConversation(conversation.convId)"
        >
          <div class="fw-semibold text-truncate">{{ getChatTitle(conversation) }}</div>
          <small class="text-muted" :class="{ 'text-white-50': selectedConversationId === conversation.convId }">
            {{ conversation.type }}
          </small>
        </button>
      </div>
    </aside>

    <section class="d-flex flex-column h-100 flex-grow-1 min-w-0">
      <template v-if="!selectedConversationId">
        <div class="d-flex flex-grow-1 align-items-center justify-content-center text-muted">
          Seleziona una chat per iniziare
        </div>
      </template>

      <template v-else>
        <header class="border-bottom p-3 bg-white">
          <h3 class="h6 mb-0">{{ getChatTitle(selectedConversation) }}</h3>
        </header>

        <div class="flex-grow-1 overflow-y-auto p-3 bg-light">
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
              <div>{{ message.text || 'Messaggio multimediale' }}</div>
              <img
                v-if="message.media?.url"
                :src="message.media.url"
                alt="media"
                class="img-fluid rounded mt-2"
                style="max-height: 220px"
              />
            </div>
          </div>
        </div>

        <div class="mt-auto border-top p-3 bg-white">
          <div class="d-flex">
            <input
              v-model="newMessageText"
              type="text"
              class="form-control me-2"
              placeholder="Scrivi un messaggio..."
              @keyup.enter="sendMessage"
            />
            <button class="btn btn-primary" @click="sendMessage">Invia</button>
          </div>
        </div>
      </template>
    </section>

    <NewChatModal @chatCreated="onChatCreated" />
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
</style>
