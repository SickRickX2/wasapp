<script setup>
defineProps({
  message: {
    type: Object,
    required: true,
  },
  userId: {
    type: String,
    default: '',
  },
  isMenuOpen: {
    type: Boolean,
    default: false,
  },
  quickEmojis: {
    type: Array,
    default: () => [],
  },
  getSenderLabel: {
    type: Function,
    required: true,
  },
  getMediaUrl: {
    type: Function,
    required: true,
  },
  aggregateReactions: {
    type: Function,
    required: true,
  },
  formatMessageTime: {
    type: Function,
    required: true,
  },
  getReplyMessageContent: {
    type: Function,
    required: true,
  },
  getReplyFromSender: {
    type: Function,
    required: true,
  },
  isMessageReadByAll: {
    type: Function,
    required: true,
  },
})

const emit = defineEmits([
  'open-image-preview',
  'toggle-menu',
  'toggle-reaction',
  'forward-message',
  'reply-message',
  'delete-message',
  'close-menu',
])
</script>

<template>
  <div
    class="d-flex mb-2"
    :class="message.sender === userId ? 'justify-content-end' : 'justify-content-start'"
  >
    <div
      class="p-2 px-3 rounded-4 text-break shadow-sm message-bubble"
      :class="{
        'bg-primary text-white': message.sender === userId && message.status !== 'deleted' && message.kind !== 'system_leave_group',
        'bg-white text-dark': message.sender !== userId && message.status !== 'deleted' && message.kind !== 'system_leave_group',
        'bg-light fst-italic': message.status === 'deleted',
        'bg-transparent text-muted text-center': message.kind === 'system_leave_group'
      }"
      :style="message.status === 'deleted' ? 'border: 2px dashed #adb5bd;' : (message.kind === 'system_leave_group' ? 'border: none; box-shadow: none;' : '')"
    >
      <div v-if="message.kind === 'system_leave_group'" class="d-flex align-items-center justify-content-center gap-2 small">
        <span class="material-symbols-outlined" style="font-size: 16px;">logout</span>
        <span>{{ message.text }} left</span>
      </div>
      <div v-else-if="message.status !== 'deleted'">
        <div v-if="message.kind === 'forwarded'" class="d-flex align-items-center gap-1 mb-1 small text-muted">
          <span class="material-symbols-outlined" style="font-size: 14px;">forward</span>
          <span>Forwarded</span>
        </div>
        <div v-if="message.replyToId" class="border-start border-3 border-secondary ps-2 mb-2 bg-light bg-opacity-50 rounded py-1">
          <small class="d-block text-muted fw-bold">{{ getReplyFromSender(message.replyToId) }}</small>
          <small class="d-block text-muted text-truncate">{{ getReplyMessageContent(message.replyToId) }}</small>
        </div>
        <div class="small mb-1 opacity-75">{{ getSenderLabel(message) }}</div>
        <div v-if="message.text">{{ message.text }}</div>
        <img
          v-if="message.media?.url"
          :src="getMediaUrl(message.media.url)"
          alt="media"
          class="img-fluid rounded mt-2 media-thumb"
          style="max-height: 220px"
          role="button"
          @click="emit('open-image-preview', message.media.url)"
        />

        <div v-if="message.reactions && message.reactions.length > 0" class="d-flex flex-wrap gap-1 mt-1">
          <span
            v-for="agg in aggregateReactions(message.reactions)"
            :key="agg.emoji"
            class="badge bg-light text-dark border shadow-sm rounded-pill d-flex align-items-center gap-1 px-2 py-1"
            :title="agg.users.join(', ')"
          >
            <span style="font-size: 0.9rem;">{{ agg.emoji }}</span>
            <span v-if="agg.count > 1" class="text-muted fw-bold" style="font-size: 0.75rem;">
              {{ agg.count }}
            </span>
          </span>
        </div>
      </div>

      <div v-else class="d-flex align-items-center gap-2 deleted-message-label">
        <span class="material-symbols-outlined" style="font-size: 16px;">block</span>
        This message has been deleted
      </div>

      <div v-if="message.status !== 'deleted' && message.kind !== 'system_leave_group'" class="d-flex align-items-center justify-content-between mt-1 gap-2">
        <div
          class="d-flex align-items-center gap-1 small"
          :class="message.sender === userId ? 'text-white-50' : 'text-muted'"
        >
          {{ formatMessageTime(message.time) }}
          <span v-if="message.sender === userId && message.status !== 'deleted'" class="ms-1">
            <span
              v-if="isMessageReadByAll(message)"
              class="material-symbols-outlined text-info"
              style="font-size: 16px; vertical-align: text-bottom;"
            >
              done_all
            </span>
            <span
              v-else
              :class="message.sender === userId ? 'material-symbols-outlined text-white-50' : 'material-symbols-outlined text-muted'"
              style="font-size: 16px; vertical-align: text-bottom;"
            >
              check
            </span>
          </span>
        </div>

        <div class="dropdown" @click.stop>
          <button
            class="btn btn-sm btn-link p-0 border-0"
            :class="message.sender === userId ? 'text-white-50' : 'text-muted'"
            type="button"
            data-bs-toggle="dropdown"
            data-bs-boundary="window"
            :aria-expanded="isMenuOpen"
            title="Message options"
            @click.stop="emit('toggle-menu', message.messageId)"
          >
            <span class="material-symbols-outlined" style="font-size: 20px; vertical-align: middle;">more_vert</span>
          </button>
          <ul class="dropdown-menu dropdown-menu-end shadow-sm" :class="{ show: isMenuOpen }">
            <li class="px-2 py-1 d-flex gap-1 justify-content-center">
              <button
                v-for="emoji in quickEmojis"
                :key="emoji"
                class="btn btn-sm btn-light rounded-circle fs-5 p-1 lh-1"
                type="button"
                @click="emit('toggle-reaction', message, emoji)"
              >
                {{ emoji }}
              </button>
            </li>
            <li>
              <button class="dropdown-item d-flex align-items-center gap-2" type="button" @click="emit('forward-message', message); emit('close-menu')" data-bs-toggle="modal" data-bs-target="#forwardModal">
                <span class="material-symbols-outlined" style="font-size: 18px;">forward</span> Forward
              </button>
            </li>
            <li>
              <button class="dropdown-item d-flex align-items-center gap-2" type="button" @click="emit('reply-message', message); emit('close-menu')">
                <span class="material-symbols-outlined" style="font-size: 18px;">reply</span> Reply
              </button>
            </li>
            <li v-if="message.sender === userId">
              <button class="dropdown-item text-danger d-flex align-items-center gap-2" type="button" @click="emit('delete-message', message.messageId); emit('close-menu')">
                <span class="material-symbols-outlined" style="font-size: 18px;">delete</span> Delete
              </button>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.message-bubble {
  max-width: 75%;
  word-break: break-word;
}

.deleted-message-label {
  color: #8f96a3;
}

.media-thumb {
  cursor: zoom-in;
}
</style>
