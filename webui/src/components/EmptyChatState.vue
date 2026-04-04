<template>
  <div class="empty-chat-state d-flex flex-column h-100 justify-content-center align-items-center bg-light text-center px-4">
    <span class="material-symbols-outlined mb-4 empty-chat-icon">
      forum
    </span>
    <h2 class="fw-light text-dark mb-3">Welcome to WASAText</h2>
    <p class="text-muted empty-chat-copy">
      Select a chat or create a new one to start
    </p>
  </div>
</template>

<style scoped>
.empty-chat-state {
  border-radius: 0.5rem;
}

.empty-chat-icon {
  font-size: 100px;
  color: #ced4da;
}

.empty-chat-copy {
  max-width: 400px;
  font-size: 0.95rem;
}
</style>
