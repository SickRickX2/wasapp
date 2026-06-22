<script setup>
import { computed } from 'vue'

const props = defineProps({
  name: {
    type: String,
    required: true,
    description: 'User identifier, used as seed for DiceBear avatars',
  },
  displayName: {
    type: String,
    default: null,
    description: 'Display name of the user, used for generating initials. If not provided, uses name.',
  },
  size: {
    type: Number,
    default: 40,
    description: 'Avatar size in pixels',
  },
  realImageUrl: {
    type: String,
    default: null,
    description: 'URL of the actual user profile picture, if available',
  },
})

const avatarUrl = computed(() => {
  if (props.realImageUrl && props.realImageUrl.trim() !== '') {
    return props.realImageUrl
  }
  // se non c'è una foto reale genera un avatar con le iniziali tramite l'api dicebear, usando displayName come seed
  const seedValue = props.displayName && props.displayName.trim() !== ''
    ? props.displayName
    : props.name
  const safeSeed = encodeURIComponent(seedValue)
  return `https://api.dicebear.com/8.x/initials/svg?seed=${safeSeed}&backgroundColor=0288d1,009688,7cb342,f57c00,e53935,8e24aa&textColor=ffffff`
})
</script>

<template>
  <img
    :src="avatarUrl"
    :alt="name"
    class="rounded-circle border border-1 border-secondary-subtle"
    :style="{ width: size + 'px', height: size + 'px', objectFit: 'cover' }"
  />
</template>

<style scoped>
img {
  display: block;
}
</style>

