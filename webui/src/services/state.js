import { reactive } from 'vue'

export const state = reactive({
  userId: null,
  userName: null,
  profilePictureUrl: localStorage.getItem('profilePictureUrl') || null,
})

// Sincronizza profilePictureUrl con localStorage quando cambia
export function updateProfilePictureUrl(url) {
  state.profilePictureUrl = url
  if (url) {
    localStorage.setItem('profilePictureUrl', url)
  } else {
    localStorage.removeItem('profilePictureUrl')
  }
}
