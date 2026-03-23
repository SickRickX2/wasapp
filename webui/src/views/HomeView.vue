<script setup>
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import axios from '../services/axios'
import { state } from '../services/state'
import NewChatModal from '../components/NewChatModal.vue'
import NewGroupModal from '../components/NewGroupModal.vue'
import UserAvatar from '../components/UserAvatar.vue'

const conversations = ref([])
const currentMessages = ref([])
const selectedConversationId = ref(null)
const newMessageText = ref('')
const userNameCache = ref({})
const userPfpCache = reactive({})
const groupPhotoCache = reactive({})
const fileInput = ref(null)
const isUploadingMedia = ref(false)
const previewImageUrl = ref('')
const showConversationInfoModal = ref(false)
const openMessageMenuId = ref(null)
const groupMemberSearchQuery = ref('')
const groupMemberSearchResults = ref([])
const isAddingGroupMember = ref(false)
const selectedMediaFile = ref(null)
const selectedMediaPreviewUrl = ref('')
const messagesContainer = ref(null)
let pollingInterval = null
let groupMemberSearchTimer = null

const selectedConversation = computed(() => {
  return conversations.value.find((c) => c.convId === selectedConversationId.value) ?? null
})

function formatDayLabel(dateLike) {
  const d = new Date(dateLike)
  if (Number.isNaN(d.getTime())) return ''
  const day = String(d.getDate()).padStart(2, '0')
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const year = d.getFullYear()
  return `${day}/${month}/${year}`
}

function formatMessageTime(dateLike) {
  const d = new Date(dateLike)
  if (Number.isNaN(d.getTime())) return ''
  const hours = String(d.getHours()).padStart(2, '0')
  const minutes = String(d.getMinutes()).padStart(2, '0')
  return `${hours}:${minutes}`
}

function getMessageBodyText(message) {
  if (!message) return ''
  if (message.status === 'deleted') {
    const deletedText = message.text || 'This message has been deleted'
    return `"${deletedText.replace(/^"|"$/g, '')}"`
  }
  return message.text || ''
}

function renderSystemMessage(message) {
  if (!message) return ''

  if (message.kind === 'system_group_created') {
    return 'The group was created'
  }

  if (message.kind === 'system_add_member') {
    const rawAdded = (message.text || '').trim()
    const addedLabel = rawAdded || getUsernameFromId(message.text) || message.text || 'someone'

    if (message.sender === state.userId) {
      return `You added ${addedLabel} to the group`
    }

    const actor = getUsernameFromId(message.sender) || message.sender || 'Someone'
    return `${actor} added ${addedLabel} to the group`
  }

  if (message.kind === 'system_leave_group') {
    const leavingUserId = message.sender
    if (leavingUserId === state.userId) {
      return 'You left the group'
    }
    const leavingRaw = (message.text || '').trim()
    const leavingLabel = leavingRaw || getUsernameFromId(leavingUserId) || leavingUserId
    return `${leavingLabel} left the group`
  }

  return message.text || ''
}

const sortedMessages = computed(() => {
  return currentMessages.value.slice().reverse()
})

const messageTimeline = computed(() => {
  const timeline = []
  let lastDay = ''

  for (const message of sortedMessages.value) {
    const dayLabel = formatDayLabel(message?.time)
    if (dayLabel && dayLabel !== lastDay) {
      timeline.push({
        itemType: 'day-separator',
        key: `day-${dayLabel}`,
        label: dayLabel,
      })
      lastDay = dayLabel
    }

    timeline.push({
      itemType: 'message',
      key: `msg-${message.messageId || message.time || timeline.length}`,
      message,
    })
  }

  return timeline
})

function getOrderedParticipantIds(conv) {
  const participants = Array.isArray(conv?.participants) ? [...conv.participants] : []
  return participants.sort((a, b) => {
    if (a === state.userId) return -1
    if (b === state.userId) return 1
    return 0
  })
}

const selectedConversationParticipantsText = computed(() => {
  const conv = selectedConversation.value
  if (!conv) return ''

  const labels = getOrderedParticipantIds(conv).map((id) => {
    if (id === state.userId) return 'Me'
    return getUsernameFromId(id) || id
  })

  return labels.join(', ')
})

const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

const isNearBottom = () => {
  if (!messagesContainer.value) return true
  const el = messagesContainer.value
  const threshold = 80
  return el.scrollHeight - el.scrollTop - el.clientHeight <= threshold
}

const getOtherParticipantId = (conv) => conv?.participants?.find((id) => id !== state.userId)

function getConversationAvatarUrl(conv) {
  if (!conv) return null
  if (conv.type === 'group') {
    const rawGroupPhoto = conv.groupPhoto || groupPhotoCache[conv.convId]
    return rawGroupPhoto ? getMediaUrl(rawGroupPhoto) : null
  }

  const otherId = getOtherParticipantId(conv)
  return otherId ? userPfpCache[otherId] ?? null : null
}

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

function toggleMessageMenu(messageId) {
  openMessageMenuId.value = openMessageMenuId.value === messageId ? null : messageId
}

function closeMessageMenu() {
  openMessageMenuId.value = null
}

function openConversationInfoModal() {
  if (!selectedConversation.value) return
  showConversationInfoModal.value = true
}

function closeConversationInfoModal() {
  showConversationInfoModal.value = false
  groupMemberSearchQuery.value = ''
  groupMemberSearchResults.value = []
}

function getParticipantAvatarUrl(userId) {
  if (!userId) return null
  if (userId === state.userId) {
    return state.profilePictureUrl || null
  }
  return userPfpCache[userId] ?? null
}

async function ensureUserPfpLoaded(userId) {
  if (!userId || userId === state.userId) return
  if (userPfpCache[userId] !== undefined) return

  userPfpCache[userId] = null
  try {
    const response = await axios.get(`/users/${userId}/pfp`)
    const pfpUrl = response.data?.pfpUrl
    userPfpCache[userId] = pfpUrl ? getMediaUrl(pfpUrl) : null
  } catch (error) {
    if (!(error?.response && error.response.status === 404)) {
      console.error(`Failed to load profile picture for user ${userId}`, error)
    }
    userPfpCache[userId] = null
  }
}

async function loadUsersPfps(userIds) {
  const ids = Array.isArray(userIds) ? userIds : []
  await Promise.all(ids.map((id) => ensureUserPfpLoaded(id)))
}

watch(showConversationInfoModal, async (isOpen) => {
  if (!isOpen) return
  const conv = selectedConversation.value
  if (!conv) return

  const participants = Array.isArray(conv.participants) ? conv.participants : []
  await Promise.all(participants.map((id) => resolveUsernameById(id)))
  await loadUsersPfps(participants)
})

watch(groupMemberSearchQuery, (value) => {
  clearTimeout(groupMemberSearchTimer)
  groupMemberSearchTimer = setTimeout(async () => {
    const conv = selectedConversation.value
    const q = value.trim()

    if (!showConversationInfoModal.value || !conv || conv.type !== 'group' || q.length === 0) {
      groupMemberSearchResults.value = []
      return
    }

    try {
      const response = await axios.get(`/users?q=${encodeURIComponent(q)}`)
      const data = response.data
      const users = Array.isArray(data?.users) ? data.users : Array.isArray(data) ? data : []
      const alreadyInGroup = new Set(conv.participants || [])
      groupMemberSearchResults.value = users.filter((u) => u?.userId && !alreadyInGroup.has(u.userId))
    } catch {
      groupMemberSearchResults.value = []
    }
  }, 300)
})

async function addUserToSelectedGroup(user) {
  const conv = selectedConversation.value
  if (!conv || conv.type !== 'group' || !user?.userId) return

  isAddingGroupMember.value = true
  try {
    await axios.post(`/conversations/${conv.convId}/participants`, {
      userIds: [user.userId],
    })

    userNameCache.value[user.userId] = user.userName || userNameCache.value[user.userId] || ''
    await ensureUserPfpLoaded(user.userId)

    groupMemberSearchQuery.value = ''
    groupMemberSearchResults.value = []

    await loadConversations()
    if (selectedConversationId.value) {
      await loadMessages(selectedConversationId.value)
    }
  } catch {
    alert('Unable to add user to group')
  } finally {
    isAddingGroupMember.value = false
  }
}

async function loadConversations() {
  try {
    if (!state.userId) return
    const response = await axios.get(`/users/${state.userId}/conversations`)
    const data = response.data
    conversations.value = Array.isArray(data?.conversations) ? data.conversations : []
    loadGroupPhotos(conversations.value)
    loadParticipantPfps(conversations.value)
    await enrichConversationNames()
  } catch (err) {
    console.error('Failed to load conversations', err)
  }
}

async function loadGroupPhotos(conversationsList) {
  for (const conv of conversationsList) {
    if (conv?.type !== 'group' || !conv?.convId) continue

    if (conv.groupPhoto) {
      groupPhotoCache[conv.convId] = conv.groupPhoto
      continue
    }

    if (groupPhotoCache[conv.convId] !== undefined) continue

    groupPhotoCache[conv.convId] = null
    try {
      const response = await axios.get(`/conversations/${conv.convId}`)
      const groupPhoto = response.data?.groupPhoto || null
      groupPhotoCache[conv.convId] = groupPhoto
    } catch (error) {
      if (!(error?.response && error.response.status === 404)) {
        console.error(`Failed to load group photo for conversation ${conv.convId}`, error)
      }
      groupPhotoCache[conv.convId] = null
    }
  }
}

async function loadParticipantPfps(conversationsList) {
  for (const conv of conversationsList) {
    if (conv?.type !== 'private') continue

    const otherId = getOtherParticipantId(conv)
    if (!otherId) continue
    await ensureUserPfpLoaded(otherId)
  }
}

async function loadMessages(conversationId, options = {}) {
  const { forceScroll = false } = options
  try {
    const shouldStickToBottom = forceScroll || isNearBottom() || currentMessages.value.length === 0
    const response = await axios.get(`/conversations/${conversationId}/messages`)
    const data = response.data
    if (Array.isArray(data?.messages)) {
      currentMessages.value = data.messages
      const senderIds = new Set(data.messages.map((m) => m?.sender).filter(Boolean))
      await Promise.all(Array.from(senderIds).map((id) => resolveUsernameById(id)))
      if (shouldStickToBottom) {
        await scrollToBottom()
      }
      return
    }
    currentMessages.value = Array.isArray(data) ? data : []
    const senderIds = new Set(currentMessages.value.map((m) => m?.sender).filter(Boolean))
    await Promise.all(Array.from(senderIds).map((id) => resolveUsernameById(id)))
    if (shouldStickToBottom) {
      await scrollToBottom()
    }
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
  await loadMessages(id, { forceScroll: true })
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
    await loadMessages(selectedConversationId.value, { forceScroll: true })
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

const reactToMessage = (msgId) => {
  console.log('Reagisci a', msgId)
}

const forwardMessage = (msgId) => {
  console.log('Inoltra', msgId)
}

async function deleteMessage(messageId) {
  if (!confirm('Eliminare il messaggio?')) return
  if (!selectedConversationId.value || !messageId) return

  try {
    const response = await axios.delete(`/conversations/${selectedConversationId.value}/messages/${messageId}`)
    const updatedMessage = response.data
    const index = currentMessages.value.findIndex((m) => m?.messageId === messageId)
    if (index < 0) return

    if (updatedMessage && updatedMessage.messageId) {
      currentMessages.value[index] = {
        ...currentMessages.value[index],
        ...updatedMessage,
      }
    } else {
      currentMessages.value[index] = {
        ...currentMessages.value[index],
        status: 'deleted',
        text: '',
        media: null,
        mediaId: '',
      }
    }
  } catch {
    alert('Error while deleting message')
  }
}

async function onChatCreated(convId) {
  await loadConversations()
  selectedConversationId.value = convId
  await loadMessages(convId, { forceScroll: true })
}

async function onGroupCreated(convId) {
  await loadConversations()

  if (convId) {
    try {
      const response = await axios.get(`/conversations/${convId}`)
      groupPhotoCache[convId] = response.data?.groupPhoto || null
    } catch {
      groupPhotoCache[convId] = groupPhotoCache[convId] || null
    }
  }

  selectedConversationId.value = convId
  await loadMessages(convId, { forceScroll: true })
}

async function leaveSelectedGroup() {
  const conv = selectedConversation.value
  if (!conv || conv.type !== 'group') return

  const confirmed = confirm(`Leave group "${conv.groupName || 'Group'}"?`)
  if (!confirmed) return

  try {
    await axios.delete(`/conversations/${conv.convId}/participants/me`)
    showConversationInfoModal.value = false
    selectedConversationId.value = null
    currentMessages.value = []
    clearSelectedMedia()
    await loadConversations()
  } catch (err) {
    if (err?.response?.status === 404) {
      alert('Group not found or you are not a participant anymore')
      showConversationInfoModal.value = false
      selectedConversationId.value = null
      currentMessages.value = []
      await loadConversations()
      return
    }
    alert('Unable to leave the group')
  }
}

onMounted(async () => {
  await loadConversations()
  pollingInterval = setInterval(syncData, 3000)
  window.addEventListener('click', closeMessageMenu)
})

onUnmounted(() => {
  if (pollingInterval) clearInterval(pollingInterval)
  if (groupMemberSearchTimer) clearTimeout(groupMemberSearchTimer)
  window.removeEventListener('click', closeMessageMenu)
})
</script>

<template>
  <div class="d-flex h-100 overflow-hidden bg-white" style="min-height: 0;">
    <aside class="d-flex flex-column h-100 border-end chat-sidebar">
      <div class="p-3 border-bottom d-flex align-items-center">
        <h2 class="h5 mb-0">Chat</h2>
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
            :name="conversation.type === 'group' ? getChatTitle(conversation) : getOtherParticipantId(conversation) || 'user'" 
            :displayName="conversation.type === 'group' ? conversation.groupName || 'Group' : getUsernameFromId(conversation.participants?.find(pid => pid !== state.userId))"
            :size="40"
            :realImageUrl="getConversationAvatarUrl(conversation)"
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
        <header
          class="border-bottom p-3 bg-white flex-shrink-0 d-flex align-items-center gap-3 chat-header-clickable"
          role="button"
          @click="openConversationInfoModal"
        >
          <UserAvatar 
            :name="selectedConversation?.type === 'group' ? getChatTitle(selectedConversation) : getOtherParticipantId(selectedConversation) || 'user'" 
            :displayName="selectedConversation?.type === 'group' ? selectedConversation?.groupName || 'Group' : getUsernameFromId(selectedConversation?.participants?.find(pid => pid !== state.userId))"
            :size="40"
            :realImageUrl="getConversationAvatarUrl(selectedConversation)"
          />
          <div class="min-w-0">
            <h3 class="h6 mb-0 text-truncate">{{ getChatTitle(selectedConversation) }}</h3>
            <small v-if="selectedConversation?.type === 'group'" class="text-muted text-truncate d-block">{{ selectedConversationParticipantsText }}</small>
          </div>
        </header>

        <div ref="messagesContainer" class="flex-grow-1 overflow-y-auto p-3" style="min-height: 0;">
          <div
            v-for="(item, index) in messageTimeline"
            :key="item.key || index"
          >
            <div v-if="item.itemType === 'day-separator'" class="d-flex justify-content-center my-3">
              <span class="date-separator-badge">{{ item.label }}</span>
            </div>

            <div
              v-else-if="item.itemType === 'message' && item.message.kind?.startsWith('system_')"
              class="d-flex justify-content-center w-100 mb-2"
            >
              <span class="system-message-chip">{{ renderSystemMessage(item.message) }}</span>
            </div>

            <div
              v-else-if="item.itemType === 'message'"
              class="d-flex mb-2"
              :class="item.message.sender === state.userId ? 'justify-content-end' : 'justify-content-start'"
            >
              <div
                class="px-3 py-2 rounded-3 message-bubble"
                :class="item.message.sender === state.userId ? 'bg-primary text-white' : 'bg-white border'"
              >
                <div class="small mb-1 opacity-75">{{ getSenderLabel(item.message) }}</div>
                <div v-if="item.message.text">{{ getMessageBodyText(item.message) }}</div>
                <img
                  v-if="item.message.media?.url"
                  :src="getMediaUrl(item.message.media.url)"
                  alt="media"
                  class="img-fluid rounded mt-2 media-thumb"
                  style="max-height: 220px"
                  role="button"
                  @click="openImagePreview(item.message.media.url)"
                />
                <div class="d-flex align-items-center justify-content-between mt-1 gap-2">
                  <div
                    class="small"
                    :class="item.message.sender === state.userId ? 'text-white-50' : 'text-muted'"
                  >
                    {{ formatMessageTime(item.message.time) }}
                  </div>
                </div>

                <div class="d-flex justify-content-start mt-1" v-if="item.message.status !== 'deleted'">
                  <div class="dropdown" @click.stop>
                    <button
                      class="btn btn-sm btn-link p-0 border-0"
                      :class="item.message.sender === state.userId ? 'text-white-50' : 'text-muted'"
                      type="button"
                      :aria-expanded="openMessageMenuId === item.message.messageId"
                      title="Opzioni messaggio"
                      @click.stop="toggleMessageMenu(item.message.messageId)"
                    >
                      <span class="material-symbols-outlined" style="font-size: 20px; vertical-align: middle;">more_vert</span>
                    </button>
                    <ul class="dropdown-menu dropdown-menu-start shadow-sm" :class="{ show: openMessageMenuId === item.message.messageId }">
                      <li>
                        <button class="dropdown-item d-flex align-items-center gap-2" type="button" @click="reactToMessage(item.message.messageId); closeMessageMenu()">
                          <span class="material-symbols-outlined" style="font-size: 18px;">add_reaction</span> Reagisci
                        </button>
                      </li>
                      <li>
                        <button class="dropdown-item d-flex align-items-center gap-2" type="button" @click="forwardMessage(item.message.messageId); closeMessageMenu()">
                          <span class="material-symbols-outlined" style="font-size: 18px;">forward</span> Inoltra
                        </button>
                      </li>
                      <li><hr class="dropdown-divider"></li>
                      <li v-if="item.message.sender === state.userId">
                        <button class="dropdown-item text-danger d-flex align-items-center gap-2" type="button" @click="deleteMessage(item.message.messageId); closeMessageMenu()">
                          <span class="material-symbols-outlined" style="font-size: 18px;">delete</span> Elimina
                        </button>
                      </li>
                    </ul>
                  </div>
                </div>
              </div>
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
    <NewGroupModal @groupCreated="onGroupCreated" />

    <div
      v-if="showConversationInfoModal && selectedConversation"
      class="conversation-info-overlay"
      @click="closeConversationInfoModal"
    >
      <div class="conversation-info-modal" @click.stop>
        <button type="button" class="btn-close conversation-info-close" aria-label="Close" @click="closeConversationInfoModal"></button>

        <UserAvatar
          :name="selectedConversation.type === 'group' ? getChatTitle(selectedConversation) : getOtherParticipantId(selectedConversation) || 'user'"
          :displayName="selectedConversation.type === 'group' ? selectedConversation.groupName || 'Group' : getUsernameFromId(getOtherParticipantId(selectedConversation))"
          :size="104"
          :realImageUrl="getConversationAvatarUrl(selectedConversation)"
          class="mx-auto mb-3"
        />

        <h4 class="h5 text-center mb-2">{{ getChatTitle(selectedConversation) }}</h4>
        <p class="text-center text-muted mb-4">{{ selectedConversation.type === 'group' ? 'Group chat' : 'Private chat' }}</p>

        <div v-if="selectedConversation.type === 'group'" class="mb-4">
          <h6 class="mb-2">Participants</h6>
          <div class="list-group participants-list">
            <div
              v-for="participantId in getOrderedParticipantIds(selectedConversation)"
              :key="participantId"
              class="list-group-item d-flex align-items-center gap-2"
            >
              <UserAvatar
                :name="participantId"
                :displayName="participantId === state.userId ? 'Me' : getUsernameFromId(participantId)"
                :size="30"
                :realImageUrl="getParticipantAvatarUrl(participantId)"
              />
              <span class="small text-truncate">{{ participantId === state.userId ? 'Me' : (getUsernameFromId(participantId) || participantId) }}</span>
            </div>
          </div>
        </div>

        <div v-if="selectedConversation.type === 'group'" class="mb-4">
          <h6 class="mb-2">Add participant</h6>
          <input
            v-model="groupMemberSearchQuery"
            type="text"
            class="form-control form-control-sm mb-2"
            placeholder="Search user..."
          />
          <div class="list-group participants-list" v-if="groupMemberSearchResults.length > 0">
            <button
              v-for="user in groupMemberSearchResults"
              :key="user.userId"
              type="button"
              class="list-group-item list-group-item-action d-flex justify-content-between align-items-center"
              :disabled="isAddingGroupMember"
              @click="addUserToSelectedGroup(user)"
            >
              <span class="text-truncate me-2">{{ user.userName }}</span>
              <span class="badge text-bg-light">Add</span>
            </button>
          </div>
        </div>

        <div v-if="selectedConversation.type === 'group'" class="d-flex justify-content-center">
          <button type="button" class="btn btn-outline-danger" @click="leaveSelectedGroup">
            Leave group
          </button>
        </div>
      </div>
    </div>

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

.date-separator-badge {
  font-size: 0.75rem;
  color: #6c757d;
  background: #eef1f4;
  border-radius: 999px;
  padding: 0.2rem 0.6rem;
}

.system-message-chip {
  font-size: 0.8rem;
  color: #6c757d;
  background: #f3f4f6;
  border: 1px solid #e5e7eb;
  border-radius: 999px;
  padding: 0.28rem 0.7rem;
}

.chat-header-clickable {
  cursor: pointer;
}

.conversation-info-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2100;
  padding: 1rem;
}

.conversation-info-modal {
  position: relative;
  width: min(92vw, 420px);
  background: #fff;
  border-radius: 0.75rem;
  padding: 1.25rem 1.25rem 1.5rem;
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.2);
}

.conversation-info-close {
  position: absolute;
  top: 0.85rem;
  right: 0.85rem;
}

.participants-list {
  max-height: 180px;
  overflow-y: auto;
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
