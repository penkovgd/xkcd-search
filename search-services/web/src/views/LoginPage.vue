<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { NForm, NFormItem, NInput, NButton } from 'naive-ui'
import apiClient from '@/utils/axios'

defineOptions({
  name: 'LoginPage',
})

const router = useRouter()
const message = useMessage()

const formRef = ref<InstanceType<typeof NForm> | null>(null)
const loading = ref(false)

const formValue = ref({
  username: '',
  password: '',
})

const rules = {
  username: {
    required: true,
    message: 'Username is required',
    trigger: ['input', 'blur'],
  },
  password: {
    required: true,
    message: 'Password is required',
    trigger: ['input', 'blur'],
  },
}

const goToHome = () => {
  router.push('/')
}

const handleLogin = async () => {
  try {
    await formRef.value?.validate()
    loading.value = true

    const response = await apiClient.post('/api/login', {
      name: formValue.value.username,
      password: formValue.value.password,
    })

    // Сохраняем токен (можно в localStorage или store)
    const token = response.data
    localStorage.setItem('token', token)

    message.success('Login successful')
    // Перенаправляем на нужную страницу после логина
    router.push('/admin')
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (error: any) {
    loading.value = false
    if (error.response?.status === 401) {
      message.error('Invalid username or password')
    } else if (error.response) {
      message.error('Login failed. Please try again.')
    } else {
      message.error('Network error. Please check your connection.')
    }
  }
}
</script>

<template>
  <div class="login-container">
    <div class="login-card surface">
      <!-- Logo -->
      <div class="logo-section">
        <img class="logo-image" src="/logo.png" alt="xkcd logo" @click="goToHome" />
      </div>

      <!-- Login Form -->
      <div class="form-section">
        <h2 class="form-title">Admin Login</h2>
        <n-form ref="formRef" :model="formValue" :rules="rules" size="large" class="login-form">
          <n-form-item label="Username" path="username">
            <n-input
              v-model:value="formValue.username"
              placeholder="Enter your username.."
              :disabled="loading"
              @keyup.enter="handleLogin"
            />
          </n-form-item>

          <n-form-item label="Password" path="password">
            <n-input
              v-model:value="formValue.password"
              placeholder="Enter your password.."
              :disabled="loading"
              @keyup.enter="handleLogin"
              type="password"
              show-password-on="click"
            >
            </n-input>
          </n-form-item>

          <n-form-item>
            <n-button
              type="primary"
              :loading="loading"
              :disabled="loading"
              @click="handleLogin"
              class="login-button"
            >
              Login
            </n-button>
          </n-form-item>
        </n-form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  /* background-color: var(--color-background); */
  padding: 4px;
  border: 1.5px solid var(--surface-border);
}

.login-card {
  display: flex;
  flex-direction: row;
  gap: 40px;
  padding: 40px;
  width: 100%;
  max-width: 700px;
}

.logo-section {
  display: flex;
  align-items: flex-start;
  flex-shrink: 0;
}

.logo-image {
  cursor: pointer;
  max-width: 200px;
  width: auto;
  height: auto;
  object-fit: contain;
}

.logo-image:hover {
  opacity: 0.8;
}

.form-section {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.form-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-dark);
  margin-bottom: 24px;
  margin-top: 0;
}

.login-form {
  width: 100%;
}

.login-button {
  width: 100%;
  background: var(--color-primary);
  border: 1.5px solid var(--surface-border);
  color: white;
  font-weight: 600;
  margin-top: 8px;
  font-variant: small-caps;
  font-family: Lucida, Helvetica, sans-serif;
  font-size: 16px;
  padding: 1.5px 12px;
  margin: 0 4px;
  text-decoration: none;
  border-radius: 3px;
  box-shadow: 0 0 5px 0 gray;
}

.login-button:hover {
  opacity: 0.9;
}

@media (max-width: 768px) {
  .login-card {
    flex-direction: column;
    gap: 30px;
    padding: 30px 20px;
  }

  .logo-section {
    justify-content: center;
  }

  .logo-image {
    max-width: 150px;
  }
}

@media (max-width: 480px) {
  .login-card {
    padding: 20px 15px;
  }

  .logo-image {
    max-width: 120px;
  }
}
</style>
