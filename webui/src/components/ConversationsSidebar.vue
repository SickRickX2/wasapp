<script setup>
import UserAvatar from './UserAvatar.vue'

defineProps({
  conversations: {
    type: Array,
    default: () => [],
  },
  selectedConversationId: {
    type: [String, Number],
    default: null,
  },
  userId: {
    type: String,
    default: '',
  },
  getChatTitle: {
    type: Function,
    required: true,
  },
  getOtherParticipantId: {
    type: Function,
    required: true,
  },
  getUsernameFromId: {
    type: Function,
    required: true,
  },
  getConversationAvatarUrl: {
    type: Function,
    required: true,
  },
  formatTime: {
    type: Function,
    required: true,
  },
})

const emit = defineEmits(['select'])

function getParticipantNameFromConversation(conversation, participantId) {
  const participants = Array.isArray(conversation?.participants) ? conversation.participants : []
  const participantNames = Array.isArray(conversation?.participantNames) ? conversation.participantNames : []
  const index = participants.findIndex((id) => id === participantId)
  if (index >= 0 && participantNames[index]) return participantNames[index]
  return participantId || 'User'
}

function getLastMessageSenderLabel(conversation, currentUserId) {
  const senderId = conversation?.lastMessage?.sender
  if (!senderId) return ''
  if (senderId === currentUserId) return 'You'
  return getParticipantNameFromConversation(conversation, senderId)
}

function getLastMessageStatusIcon(conversation, currentUserId) {
  const lastMessage = conversation?.lastMessage
  if (!lastMessage || lastMessage.sender !== currentUserId || lastMessage.status === 'deleted') return ''
  return lastMessage.status === 'seen' ? 'done_all' : 'done'
}

function getLastMessageStatusClass(conversation, currentUserId, isSelected) {
  const lastMessage = conversation?.lastMessage
  if (!lastMessage || lastMessage.sender !== currentUserId || lastMessage.status === 'deleted') return ''
  if (lastMessage.status === 'seen') return isSelected ? 'text-white' : 'text-info'
  return isSelected ? 'text-white-50' : 'text-muted'
}

function getConversationLastMessagePreview(conversation, currentUserId) {
  const lastMessage = conversation?.lastMessage
  if (!lastMessage) return ''

  const senderLabel = getLastMessageSenderLabel(conversation, currentUserId)

  if (lastMessage.status === 'deleted') {
    return senderLabel ? `${senderLabel}: Deleted message` : 'Deleted message'
  }

  if (lastMessage.kind === 'system_add_member') {
    const msg = `${lastMessage.text || 'Someone'} was added`
    return senderLabel ? `${senderLabel}: ${msg}` : msg
  }

  if (lastMessage.kind === 'system_leave_group') {
    const msg = `${lastMessage.text || 'Someone'} left`
    return senderLabel ? `${senderLabel}: ${msg}` : msg
  }

  const msg = lastMessage.text || (lastMessage.mediaId ? '📷 Image' : '')
  return senderLabel ? `${senderLabel}: ${msg}` : msg
}
</script>

<template>
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
        @click="emit('select', conversation.convId)"
      >
        <UserAvatar
          :name="conversation.type === 'group' ? getChatTitle(conversation) : getOtherParticipantId(conversation) || 'user'"
          :displayName="conversation.type === 'group' ? conversation.groupName || 'Group' : getUsernameFromId(conversation.participants?.find(pid => pid !== userId))"
          :size="40"
          :realImageUrl="getConversationAvatarUrl(conversation)"
          class="flex-shrink-0"
        />
        <div class="flex-grow-1 min-w-0">
          <div class="conversation-title-row mb-1">
            <span class="fw-bold text-truncate conversation-title">{{ getChatTitle(conversation) }}</span>
            <small
              v-if="conversation.lastMessage?.time"
              class="text-muted flex-shrink-0 text-nowrap conversation-time"
              :class="{ 'text-white-50': selectedConversationId === conversation.convId }"
            >
              {{ formatTime(conversation.lastMessage.time) }}
            </small>
          </div>

          <div class="conversation-preview-row">
            <div class="d-flex align-items-center gap-1 conversation-preview-wrap">
              <span
                v-if="getLastMessageStatusIcon(conversation, userId)"
                class="material-symbols-outlined conversation-preview-status"
                :class="getLastMessageStatusClass(conversation, userId, selectedConversationId === conversation.convId)"
              >
                {{ getLastMessageStatusIcon(conversation, userId) }}
              </span>
              <small
                class="text-muted text-truncate conversation-preview"
                :class="{ 'text-white-50': selectedConversationId === conversation.convId }"
              >
                {{ getConversationLastMessagePreview(conversation, userId) }}
              </small>
            </div>

            <span
              v-if="conversation.unreadCount && conversation.unreadCount > 0 && conversation.convId !== selectedConversationId"
              class="badge rounded-pill bg-success flex-shrink-0"
            >
              {{ conversation.unreadCount }}
            </span>
          </div>
        </div>
      </button>
    </div>
  </aside>
</template>

<style scoped>
.chat-sidebar {
  width: 32%;
  min-width: 260px;
  max-width: 420px;
}

.list-group-item.active {
  background-color: #10b981;
  border-color: #10b981;
}

.list-group-item.active:hover,
.list-group-item.active:focus {
  background-color: #0e9f6e;
  border-color: #0e9f6e;
}

.conversation-title-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.5rem;
}

.conversation-title {
  min-width: 0;
}

.conversation-time {
  font-size: 0.75rem;
}

.conversation-preview-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.5rem;
}

.conversation-preview {
  min-width: 0;
}

.conversation-preview-wrap {
  min-width: 0;
}

.conversation-preview-status {
  font-size: 14px;
  line-height: 1;
  flex-shrink: 0;
}
</style>
