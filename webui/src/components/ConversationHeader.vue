<script setup>
import UserAvatar from './UserAvatar.vue'

defineProps({
  selectedConversation: {
    type: Object,
    default: null,
  },
  chatTitle: {
    type: String,
    default: '',
  },
  participantsText: {
    type: String,
    default: '',
  },
  userId: {
    type: String,
    default: '',
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

const emit = defineEmits(['open-info'])
</script>

<template>
  <header
    class="border-bottom p-3 bg-white flex-shrink-0 d-flex align-items-center gap-3 chat-header-clickable"
    role="button"
    @click="emit('open-info')"
  >
    <UserAvatar
      :name="selectedConversation?.type === 'group' ? chatTitle : getOtherParticipantId(selectedConversation) || 'user'"
      :displayName="selectedConversation?.type === 'group' ? selectedConversation?.groupName || 'Group' : getUsernameFromId(selectedConversation?.participants?.find(pid => pid !== userId))"
      :size="40"
      :realImageUrl="getConversationAvatarUrl(selectedConversation)"
    />
    <div class="min-w-0">
      <h3 class="h6 mb-0 text-truncate">{{ chatTitle }}</h3>
      <small v-if="selectedConversation?.type === 'group'" class="text-muted text-truncate d-block">{{ participantsText }}</small>
    </div>
  </header>
</template>

<style scoped>
.chat-header-clickable {
  cursor: pointer;
}
</style>
