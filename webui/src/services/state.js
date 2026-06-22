import { reactive } from 'vue'

// stato globale reattivo condiviso tra tutti i componenti, inizializzato da localStorage
export const state = reactive({
  userId: null,
  userName: null,
  profilePictureUrl: localStorage.getItem('profilePictureUrl') || null,
})

// sincronizza profilePictureUrl con localStorage quando cambia
export function updateProfilePictureUrl(url) {
  state.profilePictureUrl = url
  if (url) {
    localStorage.setItem('profilePictureUrl', url)
  } else {
    localStorage.removeItem('profilePictureUrl')
  }
}
