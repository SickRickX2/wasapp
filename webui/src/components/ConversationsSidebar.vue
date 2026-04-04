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

function getConversationLastMessagePreview(conversation) {
  const lastMessage = conversation?.lastMessage
  if (!lastMessage) return ''

  if (lastMessage.status === 'deleted') {
    return 'Deleted message'
  }

  if (lastMessage.kind === 'system_add_member') {
    return `${lastMessage.text || 'Someone'} was added`
  }

  if (lastMessage.kind === 'system_leave_group') {
    return `${lastMessage.text || 'Someone'} left`
  }

  return lastMessage.text || (lastMessage.mediaId ? '📷 Image' : '')
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
          <div class="d-flex justify-content-between align-items-center mb-1 gap-2">
            <span class="fw-bold text-truncate">{{ getChatTitle(conversation) }}</span>
            <small
              v-if="conversation.lastMessage?.time"
              class="text-muted flex-shrink-0"
              style="font-size: 0.75rem;"
              :class="{ 'text-white-50': selectedConversationId === conversation.convId }"
            >
              {{ formatTime(conversation.lastMessage.time) }}
            </small>
          </div>

          <div class="d-flex justify-content-between align-items-center gap-2">
            <small
              class="text-muted text-truncate pe-2 flex-grow-1"
              style="max-width: 85%;"
              :class="{ 'text-white-50': selectedConversationId === conversation.convId }"
            >
              {{ getConversationLastMessagePreview(conversation) }}
            </small>

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
</style>
