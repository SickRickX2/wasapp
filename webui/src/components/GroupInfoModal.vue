<script setup>
import { ref } from 'vue'
import UserAvatar from './UserAvatar.vue'

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  conversation: {
    type: Object,
    default: null,
  },
  chatTitle: {
    type: String,
    default: '',
  },
  conversationAvatarName: {
    type: String,
    default: '',
  },
  conversationAvatarDisplayName: {
    type: String,
    default: '',
  },
  conversationAvatarUrl: {
    type: String,
    default: null,
  },
  participants: {
    type: Array,
    default: () => [],
  },
  groupMemberSearchQuery: {
    type: String,
    default: '',
  },
  groupMemberSearchResults: {
    type: Array,
    default: () => [],
  },
  isAddingGroupMember: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits([
  'close',
  'update:groupMemberSearchQuery',
  'add-user',
  'leave-group',
  'update-name',
  'update-photo',
])

const isEditingName = ref(false)
const editNameValue = ref(props.chatTitle)
const fileInput = ref(null)

function onSearchInput(event) {
  emit('update:groupMemberSearchQuery', event.target.value)
}

function triggerPhotoUpload() {
  fileInput.value?.click()
}

function onFileSelected(event) {
  const file = event.target.files?.[0]
  if (file) emit('update-photo', file)
  event.target.value = ''
}

function saveGroupName() {
  if (editNameValue.value.trim()) {
    emit('update-name', editNameValue.value)
    isEditingName.value = false
  }
}

function close() {
  emit('close')
}
</script>

<template>
  <div
    v-if="show && conversation"
    class="conversation-info-overlay"
    @click="close"
  >
    <div class="conversation-info-modal" @click.stop>
      <button type="button" class="btn-close conversation-info-close" aria-label="Close" @click="close"></button>

      <input type="file" ref="fileInput" class="d-none" accept="image/*" @change="onFileSelected">

      <div class="position-relative d-inline-block text-center w-100 mb-3">
        <UserAvatar
          :name="conversationAvatarName"
          :displayName="conversationAvatarDisplayName"
          :size="104"
          :realImageUrl="conversationAvatarUrl"
          class="mx-auto"
        />
        <button v-if="conversation.type === 'group'" class="btn btn-sm btn-light position-absolute bottom-0 start-50 translate-middle-x border shadow-sm rounded-pill" style="font-size: 0.75rem;" @click="triggerPhotoUpload"><span class="material-symbols-outlined" style="font-size: 14px; vertical-align: text-bottom;">edit</span> Cambia</button>
      </div>

      <div v-if="conversation.type === 'group'" class="mb-2 text-center">
        <div v-if="!isEditingName" class="d-flex align-items-center justify-content-center gap-2">
          <h4 class="h5 mb-0">{{ chatTitle }}</h4>
          <button class="btn btn-link text-muted p-0" @click="isEditingName = true; editNameValue = chatTitle"><span class="material-symbols-outlined" style="font-size: 18px;">edit</span></button>
        </div>
        <div v-else class="d-flex align-items-center justify-content-center gap-1 px-3">
          <input type="text" class="form-control form-control-sm" v-model="editNameValue" @keyup.enter="saveGroupName">
          <button class="btn btn-sm btn-primary" @click="saveGroupName"><span class="material-symbols-outlined" style="font-size: 16px;">check</span></button>
          <button class="btn btn-sm btn-outline-secondary" @click="isEditingName = false"><span class="material-symbols-outlined" style="font-size: 16px;">close</span></button>
        </div>
      </div>
      <h4 v-else class="h5 text-center mb-2">{{ chatTitle }}</h4>
      <p class="text-center text-muted mb-4">{{ conversation.type === 'group' ? 'Group chat' : 'Private chat' }}</p>

      <div v-if="conversation.type === 'group'" class="mb-4">
        <h6 class="mb-2">Participants</h6>
        <div class="list-group participants-list">
          <div
            v-for="participant in participants"
            :key="participant.id"
            class="list-group-item d-flex align-items-center gap-2"
          >
            <UserAvatar
              :name="participant.id"
              :displayName="participant.label"
              :size="30"
              :realImageUrl="participant.avatarUrl"
            />
            <span class="small text-truncate">{{ participant.label }}</span>
          </div>
        </div>
      </div>

      <div v-if="conversation.type === 'group'" class="mb-4">
        <h6 class="mb-2">Add participant</h6>
        <input
          :value="groupMemberSearchQuery"
          type="text"
          class="form-control form-control-sm mb-2"
          placeholder="Search user..."
          @input="onSearchInput"
        />
        <div class="list-group participants-list" v-if="groupMemberSearchResults.length > 0">
          <button
            v-for="user in groupMemberSearchResults"
            :key="user.userId"
            type="button"
            class="list-group-item list-group-item-action d-flex justify-content-between align-items-center"
            :disabled="isAddingGroupMember"
            @click="emit('add-user', user)"
          >
            <span class="text-truncate me-2">{{ user.userName }}</span>
            <span class="badge text-bg-light">Add</span>
          </button>
        </div>
      </div>

      <div v-if="conversation.type === 'group'" class="d-flex justify-content-center">
        <button type="button" class="btn btn-outline-danger" @click="emit('leave-group')">
          Leave group
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
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
</style>
