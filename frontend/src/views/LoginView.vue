<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const accountId = ref('')
const error = ref('')

const handleLogin = () => {
  if (!accountId.value) {
    error.value = 'Por favor, insira o ID da conta.'
    return
  }
  // Simular login salvando ID no localStorage
  localStorage.setItem('account_id', accountId.value)
  router.push('/dashboard')
}
</script>

<template>
  <div class="login-container">
    <div class="card login-card fade-in">
      <div class="logo">
        <h1 class="heading text-gradient">Acme Inc.</h1>
        <p>O futuro do seu dinheiro.</p>
      </div>
      
      <form @submit.prevent="handleLogin" class="login-form">
        <div class="form-group">
          <label class="label">ID da Conta</label>
          <input 
            v-model="accountId" 
            type="text" 
            class="input" 
            placeholder="Ex: 550e8400-e29b-41d4-a716-446655440000"
          >
        </div>
        
        <p v-if="error" class="error-msg">{{ error }}</p>
        
        <button type="submit" class="btn btn-primary full-width">
          Acessar Conta
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: radial-gradient(circle at top right, #1e1b4b, var(--background));
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: rgba(30, 41, 59, 0.7);
  backdrop-filter: blur(10px);
}

.logo {
  text-align: center;
  margin-bottom: 2rem;
}

.logo h1 {
  font-size: 2.5rem;
  margin-bottom: 0.5rem;
}

.logo p {
  color: var(--text-muted);
}

.form-group {
  margin-bottom: 1.5rem;
}

.full-width {
  width: 100%;
}

.error-msg {
  color: var(--danger);
  font-size: 0.875rem;
  margin-bottom: 1rem;
  text-align: center;
}
</style>
