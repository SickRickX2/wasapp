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
})

const emit = defineEmits(['select'])
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
          <div class="fw-semibold text-truncate">{{ getChatTitle(conversation) }}</div>
          <small class="text-muted" :class="{ 'text-white-50': selectedConversationId === conversation.convId }">
            {{ conversation.type }}
          </small>
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
</style>
