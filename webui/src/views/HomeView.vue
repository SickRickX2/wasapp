<script setup>
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import axios from '../services/axios'
import { state } from '../services/state'
import NewChatModal from '../components/NewChatModal.vue'
import NewGroupModal from '../components/NewGroupModal.vue'
import ConversationsSidebar from '../components/ConversationsSidebar.vue'
import EmptyChatState from '../components/EmptyChatState.vue'
import MessagesLoadingState from '../components/MessagesLoadingState.vue'
import ConversationHeader from '../components/ConversationHeader.vue'
import GroupInfoModal from '../components/GroupInfoModal.vue'
import MessageBubble from '../components/MessageBubble.vue'

const conversations = ref([])
const currentMessages = ref([])
const selectedConversationId = ref(null)
const newMessageText = ref('')
const userNameCache = ref({})
const userPfpCache = reactive({})
const groupPhotoCache = reactive({})
const fileInput = ref(null)
const isUploadingMedia = ref(false)
const isLoadingMessages = ref(false)
const isLoadingMoreMessages = ref(false)
const previewImageUrl = ref('')
const showConversationInfoModal = ref(false)
const openMessageMenuId = ref(null)
const groupMemberSearchQuery = ref('')
const groupMemberSearchResults = ref([])
const isAddingGroupMember = ref(false)
const selectedMediaFile = ref(null)
const selectedMediaPreviewUrl = ref('')
const messagesContainer = ref(null)
const currentOffset = ref(0)
const hasMoreMessages = ref(true)
const messageToForward = ref(null)
const messageToReply = ref(null)
const quickEmojis = ['👍', '❤️', '😂', '😯', '😢', '🙏']
const isPollingInProgress = ref(false)
let pollingInterval = null
let groupMemberSearchTimer = null

// Anti-crash blindatura: traccia quali messaggi sono già stati processati
const locallyMarkedAsSeen = new Set()

const selectedConversation = computed(() => {
  return conversations.value.find((c) => c.convId === selectedConversationId.value) ?? null
})

const sortedConversationsList = computed(() => {
  if (!conversations.value) return []

  return [...conversations.value].sort((a, b) => {
    const timeA = a?.lastMessage?.time ? new Date(a.lastMessage.time).getTime() : 0
    const timeB = b?.lastMessage?.time ? new Date(b.lastMessage.time).getTime() : 0
    return timeB - timeA
  })
})

function formatDayLabel(dateLike) {
  const d = new Date(dateLike)
  if (Number.isNaN(d.getTime())) return ''
  const day = String(d.getDate()).padStart(2, '0')
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const year = d.getFullYear()
  return `${day}/${month}/${year}`
}

function formatTime(dateLike) {
  const d = new Date(dateLike)
  if (Number.isNaN(d.getTime())) return ''
  const hours = String(d.getHours()).padStart(2, '0')
  const minutes = String(d.getMinutes()).padStart(2, '0')
  return `${hours}:${minutes}`
}

function formatMessageTime(dateLike) {
  return formatTime(dateLike)
}

const aggregateReactions = (reactions) => {
  if (!reactions || reactions.length === 0) return []

  if (Array.isArray(reactions) && reactions.some((r) => Array.isArray(r?.users))) {
    return reactions.map((r) => ({
      emoji: r.emoji,
      count: Array.isArray(r.users) ? r.users.length : 0,
      users: Array.isArray(r.users) ? r.users : [],
      userIds: Array.isArray(r.userIds) ? r.userIds : [],
    }))
  }

  const counts = {}
  reactions.forEach((r) => {
    if (!counts[r.emoji]) {
      counts[r.emoji] = { emoji: r.emoji, count: 0, users: [], userIds: [] }
    }
    counts[r.emoji].count++
    counts[r.emoji].users.push(r.userId)
    counts[r.emoji].userIds.push(r.userId)
  })
  return Object.values(counts)
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
      key: `msg-${message.messageId || message.time || 'fallback'}`,
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

const selectedConversationModalParticipants = computed(() => {
  const conv = selectedConversation.value
  if (!conv || conv.type !== 'group') return []

  return getOrderedParticipantIds(conv).map((participantId) => ({
    id: participantId,
    label: participantId === state.userId ? 'Me' : getUsernameFromId(participantId) || participantId,
    avatarUrl: getParticipantAvatarUrl(participantId),
  }))
})

const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTo({
      top: messagesContainer.value.scrollHeight,
      behavior: 'smooth',
    })
  }
}

const isNearBottom = () => {
  if (!messagesContainer.value) return true
  const el = messagesContainer.value
  const threshold = 80
  return el.scrollHeight - el.scrollTop - el.clientHeight <= threshold
}

watch(
  () => currentMessages.value?.length,
  (newLength, oldLength) => {
    if (isLoadingMoreMessages.value) return
    if (newLength > (oldLength || 0)) {
      scrollToBottom()
    }
  }
)

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

function getReplyMessageContent(messageId) {
  if (!messageId) return ''
  const repliedMessage = currentMessages.value.find(m => m.messageId === messageId)
  if (!repliedMessage) return ''
  if (repliedMessage.media?.url) return '📷 Photo'
  return repliedMessage.text || ''
}

function getReplyFromSender(messageId) {
  if (!messageId) return ''
  const repliedMessage = currentMessages.value.find(m => m.messageId === messageId)
  if (!repliedMessage) return ''
  return getSenderLabel(repliedMessage)
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

async function ensureUserPfpLoaded(userId, forceRefresh = false) {
  if (!userId || userId === state.userId) return
  if (!forceRefresh && userPfpCache[userId] !== undefined) return

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
  await Promise.all(ids.map((id) => ensureUserPfpLoaded(id, true)))
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
    await ensureUserPfpLoaded(user.userId, true)

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
    const nextConversations = Array.isArray(data?.conversations) ? [...data.conversations] : []
    conversations.value = nextConversations
    loadGroupPhotos(nextConversations)
    loadParticipantPfps(nextConversations)
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
    await ensureUserPfpLoaded(otherId, true)
  }
}

async function loadMessages(conversationId, options = {}) {
  const { forceScroll = false, isPolling = false, showLoading = false } = options
  
  if (showLoading && !isPolling) {
    isLoadingMessages.value = true
  }
  try {
    const shouldStickToBottom = forceScroll || isNearBottom() || currentMessages.value.length === 0
    const response = await axios.get(`/conversations/${conversationId}/messages?limit=20&offset=0`)
    const data = response.data
    if (Array.isArray(data?.messages)) {
      currentMessages.value = data.messages
      hasMoreMessages.value = data.messages.length >= 20
      const senderIds = new Set(data.messages.map((m) => m?.sender).filter(Boolean))
      await Promise.all(Array.from(senderIds).map((id) => resolveUsernameById(id)))
      if (shouldStickToBottom) {
        await scrollToBottom()
      }
      await markUnreadMessagesAsSeen(currentMessages.value)
      return
    }
    currentMessages.value = Array.isArray(data) ? data : []
    hasMoreMessages.value = currentMessages.value.length >= 20
    const senderIds = new Set(currentMessages.value.map((m) => m?.sender).filter(Boolean))
    await Promise.all(Array.from(senderIds).map((id) => resolveUsernameById(id)))
    if (shouldStickToBottom) {
      await scrollToBottom()
    }
    await markUnreadMessagesAsSeen(currentMessages.value)
  } catch (err) {
    console.error('Failed to load messages', err)
  } finally {
    if (showLoading && !isPolling) {
      isLoadingMessages.value = false
    }
  }
}

async function loadMoreMessages() {
  if (!selectedConversationId.value || !hasMoreMessages.value || isLoadingMoreMessages.value) return

  isLoadingMoreMessages.value = true
  try {
    const nextOffset = currentOffset.value + 20
    console.log('Loading more messages with offset:', nextOffset)
    const response = await axios.get(`/conversations/${selectedConversationId.value}/messages?limit=20&offset=${nextOffset}`)
    console.log('Response data:', response.data)
    const incoming = Array.isArray(response.data?.messages)
      ? response.data.messages
      : (Array.isArray(response.data) ? response.data : [])
    
    console.log('Incoming messages:', incoming.length)

    if (incoming.length < 20) {
      hasMoreMessages.value = false
    }

    // Aggiungi alla FINE, così dopo il reverse appariranno SOPRA (messaggi più vecchi in alto)
    currentMessages.value = [...currentMessages.value, ...incoming]
    currentOffset.value += 20

    const senderIds = new Set(incoming.map((m) => m?.sender).filter(Boolean))
    await Promise.all(Array.from(senderIds).map((id) => resolveUsernameById(id)))

    // Scroll giù di poco per mostrare ai messaggi nuovi senza perdere il contesto
    await nextTick()
    if (messagesContainer.value) {
      const scrollIncrement = 150 // Scroll di 150px verso il basso per mostrare i nuovi messaggi
      messagesContainer.value.scrollTop += scrollIncrement
      console.log('Scrolled down by', scrollIncrement, 'px')
    }
  } catch (err) {
    console.error('Failed to load more messages', err)
    alert(`Errore nel caricamento: ${err?.response?.status} ${err?.message}`)
  } finally {
    isLoadingMoreMessages.value = false
  }
}

const isMessageReadByAll = (message) => {
  if (!selectedConversation.value) return false

  // Se è una chat privata, basta lo status standard
  if (selectedConversation.value.type !== 'group') {
    return message.status === 'seen'
  }

  // SE È UN GRUPPO:
  // Se il backend fornisce un array (es. message.readBy), calcola la lunghezza
  if (Array.isArray(message.readBy)) {
    const expectedReaders = (selectedConversation.value.participants?.length || 2) - 1
    return message.readBy.length >= expectedReaders
  }

  // Se il backend fornisce solo uno status globale, ci fidiamo di quello
  return message.status === 'seen'
}

async function markUnreadMessagesAsSeen(messages) {
  if (!messages || messages.length === 0) return

  // Filtra solo i messaggi: non miei, non ancora letti, e MAI processati prima
  const unreadMessages = messages.filter(m => {
    const isMine = m.sender === state.userId
    const isUnread = m.status !== 'seen'
    const notProcessed = !locallyMarkedAsSeen.has(m.messageId)
    return !isMine && isUnread && notProcessed
  })

  for (const msg of unreadMessages) {
    // Aggiungi SUBITO al set per prevenire loop causati dal polling veloce
    locallyMarkedAsSeen.add(msg.messageId)
    try {
      await axios.put(`/conversations/${selectedConversationId.value}/messages/${msg.messageId}/seen`)
    } catch (error) {
      // Se fallisce, rimuovi dal set per riprovare al prossimo giro
      locallyMarkedAsSeen.delete(msg.messageId)
      console.error('Error marking message as seen', error)
    }
  }
}

async function syncData() {
  if (selectedConversationId.value) {
    // Se l'utente ha caricato messaggi precedenti, controlla solo i nuovi (offset=0)
    if (currentOffset.value > 0) {
      try {
        const response = await axios.get(`/conversations/${selectedConversationId.value}/messages?limit=20&offset=0`)
        const latestMessages = Array.isArray(response.data?.messages) ? response.data.messages : (Array.isArray(response.data) ? response.data : [])
        
        if (latestMessages.length > 0) {
          // Prendi solo i messaggi che NON sono già nell'array
          const existingIds = new Set(currentMessages.value.map(m => m.messageId))
          const newMessages = latestMessages.filter(m => !existingIds.has(m.messageId))
          
          if (newMessages.length > 0) {
            console.log('Found', newMessages.length, 'new messages during polling')
            // Aggiungi alla fine e scrolla automaticamente
            currentMessages.value = [...currentMessages.value, ...newMessages]
            await scrollToBottom()
          }
        }
      } catch (err) {
        console.error('Failed to check for new messages during polling', err)
      }
    } else {
      // Carica normalmente se l'utente sta guardando i messaggi recenti
      await loadMessages(selectedConversationId.value, { isPolling: true })
    }
    await markUnreadMessagesAsSeen(currentMessages.value)
  }
}

async function refreshConversationsNow() {
  await loadConversations()
}

async function selectConversation(id) {
  clearSelectedMedia()
  selectedConversationId.value = id
  currentOffset.value = 0
  hasMoreMessages.value = true
  await loadMessages(id, { forceScroll: true, showLoading: true })
  await scrollToBottom()
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
  if (text.length > 150) return
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
    if (messageToReply.value) {
      payload.replyToId = messageToReply.value.messageId
    }

    await axios.post(`/conversations/${selectedConversationId.value}/messages`, payload)
    newMessageText.value = ''
    clearSelectedMedia()
    messageToReply.value = null
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

async function toggleReaction(message, emoji) {
  if (!selectedConversationId.value || !message?.messageId || !emoji) return

  const getUserIds = (reaction) => Array.isArray(reaction?.userIds)
    ? reaction.userIds
    : reaction?.userId
      ? [reaction.userId]
      : []

  const getUsers = (reaction) => Array.isArray(reaction?.users) ? reaction.users : []

  const cloneReactions = (reactions) => (Array.isArray(reactions) ? reactions.map((reaction) => ({
    emoji: reaction.emoji,
    users: getUsers(reaction),
    userIds: getUserIds(reaction),
  })) : [])

  const reactions = Array.isArray(message.reactions) ? message.reactions : []
  const existingReactionIndex = reactions.findIndex((r) => getUserIds(r).includes(state.userId))
  const existingReaction = existingReactionIndex !== -1 ? reactions[existingReactionIndex] : null
  const isRemoving = existingReaction?.emoji === emoji

  const backupReactions = cloneReactions(reactions)
  const userDisplayName = state.userName || 'Me'

  if (!message.reactions) message.reactions = []

  const removeUserFromReaction = (reaction) => {
    reaction.userIds = getUserIds(reaction).filter((id) => id !== state.userId)
    reaction.users = getUsers(reaction).filter((name) => name !== userDisplayName)
  }

  const addUserToReaction = (reaction) => {
    if (!getUserIds(reaction).includes(state.userId)) {
      reaction.userIds = [...getUserIds(reaction), state.userId]
    }
    if (!getUsers(reaction).includes(userDisplayName)) {
      reaction.users = [...getUsers(reaction), userDisplayName]
    }
  }

  if (isRemoving) {
    const reaction = message.reactions[existingReactionIndex]
    removeUserFromReaction(reaction)
    if (getUserIds(reaction).length === 0) {
      message.reactions.splice(existingReactionIndex, 1)
    }
  } else if (existingReactionIndex !== -1) {
    const oldReaction = message.reactions[existingReactionIndex]
    removeUserFromReaction(oldReaction)
    if (getUserIds(oldReaction).length === 0) {
      message.reactions.splice(existingReactionIndex, 1)
    }

    const targetIndex = message.reactions.findIndex((r) => r.emoji === emoji)
    if (targetIndex !== -1) {
      addUserToReaction(message.reactions[targetIndex])
    } else {
      message.reactions.push({
        emoji,
        users: [userDisplayName],
        userIds: [state.userId],
      })
    }
  } else {
    const targetIndex = message.reactions.findIndex((r) => r.emoji === emoji)
    if (targetIndex !== -1) {
      addUserToReaction(message.reactions[targetIndex])
    } else {
      message.reactions.push({
        emoji,
        users: [userDisplayName],
        userIds: [state.userId],
      })
    }
  }

  closeMessageMenu()

  // 4. Esegui la chiamata API in background
  try {
    if (isRemoving) {
      await axios.delete(
        `/conversations/${selectedConversationId.value}/messages/${message.messageId}/reaction`
      )
    } else {
      await axios.put(
        `/conversations/${selectedConversationId.value}/messages/${message.messageId}/reaction`,
        { emoji }
      )
    }
  } catch {
    console.error('Errore reazione, ripristino stato...')
    // Ripristina in caso di errore di rete
    message.reactions = backupReactions
  }
}

const forwardMessage = (message) => {
  messageToForward.value = message
}

const prepareReply = (message) => {
  messageToReply.value = message
  document.querySelector('.chat-input-area input')?.focus()
}

const executeForward = async (targetConvId) => {
  if (!messageToForward.value) return
  try {
    await axios.post(
      `/conversations/${selectedConversationId.value}/messages/${messageToForward.value.messageId}/forwarded`,
      { destinationConversationId: targetConvId }
    )

    document.getElementById('closeForwardModalBtn')?.click()

    selectedConversationId.value = targetConvId
    await loadMessages(targetConvId, { forceScroll: true })
    messageToForward.value = null
  } catch (error) {
    console.error('Error forwarding message', error)
    alert('Unable to forward message')
  }
}

async function deleteMessage(messageId) {
  if (!confirm('Delete this message?')) return
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

async function handleUpdateGroupName(newName) {
  if (!selectedConversationId.value) return

  try {
    await axios.put(`/conversations/${selectedConversationId.value}/group_name`, {
      groupName: newName,
    })

    await loadConversations()
  } catch {
    alert('Unable to update group name')
  }
}

async function handleUpdateGroupPhoto(file) {
  if (!selectedConversationId.value || !file) return

  const formData = new FormData()
  formData.append('file', file)

  try {
    await axios.put(`/conversations/${selectedConversationId.value}/group_photo`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })

    await loadConversations()
  } catch {
    alert('Unable to update group photo')
  }
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
  pollingInterval = setInterval(async () => {
    if (isPollingInProgress.value) return
    isPollingInProgress.value = true
    try {
      await loadConversations()
      if (selectedConversationId.value) {
        await syncData()
      }
    } catch (e) {
      console.error('Errore nel polling:', e)
    } finally {
      isPollingInProgress.value = false
    }
  }, 3000)
  window.addEventListener('click', closeMessageMenu)
  window.addEventListener('refresh-conversations', refreshConversationsNow)
})

onUnmounted(() => {
  if (pollingInterval) clearInterval(pollingInterval)
  if (groupMemberSearchTimer) clearTimeout(groupMemberSearchTimer)
  window.removeEventListener('click', closeMessageMenu)
  window.removeEventListener('refresh-conversations', refreshConversationsNow)
})
</script>

<template>
  <div class="d-flex h-100 overflow-hidden bg-white" style="min-height: 0;">
    <ConversationsSidebar
      :conversations="sortedConversationsList"
      :selectedConversationId="selectedConversationId"
      :userId="state.userId"
      :getChatTitle="getChatTitle"
      :getOtherParticipantId="getOtherParticipantId"
      :getUsernameFromId="getUsernameFromId"
      :getConversationAvatarUrl="getConversationAvatarUrl"
      :formatTime="formatTime"
      @select="selectConversation"
    />

    <section class="d-flex flex-column h-100 min-w-0 flex-grow-1" style="min-height: 0;">
      <template v-if="!selectedConversationId">
        <EmptyChatState />
      </template>

      <template v-else>
        <ConversationHeader
          :selectedConversation="selectedConversation"
          :chatTitle="getChatTitle(selectedConversation)"
          :participantsText="selectedConversationParticipantsText"
          :userId="state.userId"
          :getOtherParticipantId="getOtherParticipantId"
          :getUsernameFromId="getUsernameFromId"
          :getConversationAvatarUrl="getConversationAvatarUrl"
          @open-info="openConversationInfoModal"
        />

        <div ref="messagesContainer" class="chat-messages-area flex-grow-1 overflow-y-auto p-3" style="min-height: 0; background-color: #efeae2;">
          <MessagesLoadingState v-if="isLoadingMessages" />
          <div v-else>
            <div v-if="hasMoreMessages" class="text-center my-3">
              <button @click="loadMoreMessages" class="btn btn-sm btn-outline-primary rounded-pill px-3" :disabled="isLoadingMoreMessages">
                {{ isLoadingMoreMessages ? 'Loading...' : 'Load previous messages' }}
              </button>
            </div>
            <div
              v-for="item in messageTimeline"
              :key="item.key"
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
              >
                <MessageBubble
                  :message="item.message"
                  :userId="state.userId"
                  :isMenuOpen="openMessageMenuId === item.message.messageId"
                  :quickEmojis="quickEmojis"
                  :getSenderLabel="getSenderLabel"
                  :getMediaUrl="getMediaUrl"
                  :aggregateReactions="aggregateReactions"
                  :formatMessageTime="formatMessageTime"
                  :getReplyMessageContent="getReplyMessageContent"
                  :getReplyFromSender="getReplyFromSender"
                  :isMessageReadByAll="isMessageReadByAll"
                  @open-image-preview="openImagePreview"
                  @toggle-menu="toggleMessageMenu"
                  @toggle-reaction="toggleReaction"
                  @forward-message="forwardMessage"
                  @reply-message="prepareReply"
                  @delete-message="deleteMessage"
                  @close-menu="closeMessageMenu"
                />
              </div>
            </div>
          </div>
        </div>

        <div class="p-3 bg-light border-top mt-auto flex-shrink-0">
          <div v-if="messageToReply" class="bg-light border-start border-4 border-primary p-2 mb-2 rounded d-flex justify-content-between align-items-center shadow-sm">
            <div class="flex-grow-1 min-w-0">
              <small class="d-block text-muted fw-bold">Reply to {{ getSenderLabel(messageToReply) }}</small>
              <small class="d-block text-truncate text-muted">{{ messageToReply.text || '(No text)' }}</small>
            </div>
            <button type="button" class="btn btn-sm btn-close ms-2" @click="messageToReply = null"></button>
          </div>
          <div class="input-group">
            <button class="btn btn-outline-secondary d-flex align-items-center" type="button" @click="triggerFileInput">
              <span class="material-symbols-outlined">add_photo_alternate</span>
            </button>
            <input type="text" class="form-control" placeholder="Type a message..." v-model="newMessageText" maxlength="150" @keyup.enter="sendMessage">
            <button class="btn btn-primary" type="button" @click="sendMessage">Send</button>
          </div>
          <div class="d-flex justify-content-end mt-1">
            <small class="text-muted ms-2" style="font-size: 0.75rem;">
              {{ newMessageText.length }}/150
            </small>
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

    <GroupInfoModal
      :show="showConversationInfoModal"
      :conversation="selectedConversation"
      :chatTitle="getChatTitle(selectedConversation)"
      :conversationAvatarName="selectedConversation?.type === 'group' ? getChatTitle(selectedConversation) : getOtherParticipantId(selectedConversation) || 'user'"
      :conversationAvatarDisplayName="selectedConversation?.type === 'group' ? selectedConversation?.groupName || 'Group' : getUsernameFromId(getOtherParticipantId(selectedConversation))"
      :conversationAvatarUrl="getConversationAvatarUrl(selectedConversation)"
      :participants="selectedConversationModalParticipants"
      :groupMemberSearchQuery="groupMemberSearchQuery"
      :groupMemberSearchResults="groupMemberSearchResults"
      :isAddingGroupMember="isAddingGroupMember"
      @close="closeConversationInfoModal"
      @update:groupMemberSearchQuery="groupMemberSearchQuery = $event"
      @update-name="handleUpdateGroupName"
      @update-photo="handleUpdateGroupPhoto"
      @add-user="addUserToSelectedGroup"
      @leave-group="leaveSelectedGroup"
    />

    <div v-if="previewImageUrl" class="image-preview-overlay" @click="closeImagePreview">
      <button class="btn btn-light image-preview-close" type="button" @click.stop="closeImagePreview">✕</button>
      <img :src="previewImageUrl" alt="Image preview" class="image-preview-full" @click.stop />
    </div>

    <div class="modal fade" id="forwardModal" tabindex="-1" aria-labelledby="forwardModalLabel" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h1 class="modal-title fs-5" id="forwardModalLabel">Forward message to...</h1>
            <button type="button" class="btn-close" id="closeForwardModalBtn" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body p-0">
            <div class="list-group list-group-flush">
              <template v-for="conv in conversations" :key="conv.convId">
                <button
                  type="button"
                  class="list-group-item list-group-item-action d-flex align-items-center justify-content-between p-3"
                  @click="executeForward(conv.convId)"
                >
                  <div class="d-flex align-items-center gap-3">
                    <UserAvatar
                      :name="conv.type === 'group' ? getChatTitle(conv) : getOtherParticipantId(conv) || 'user'"
                      :displayName="conv.type === 'group' ? conv.groupName || 'Group' : getUsernameFromId(conv.participants?.find(pid => pid !== state.userId))"
                      :size="40"
                      :realImageUrl="getConversationAvatarUrl(conv)"
                    />
                    <span class="fw-medium">{{ getChatTitle(conv) }}</span>
                  </div>
                  <span class="material-symbols-outlined text-primary">send</span>
                </button>
              </template>
              <div v-if="conversations.length === 0" class="p-4 text-center text-muted">
                No chats available for forwarding.
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
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

.material-symbols-outlined {
  font-variation-settings: 'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24;
  font-size: 24px;
  vertical-align: middle;
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
