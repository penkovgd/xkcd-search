<!-- eslint-disable @typescript-eslint/no-explicit-any -->
<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { NButton, NGrid, NGi, NPopconfirm, useMessage } from 'naive-ui'
import apiClient from '@/utils/axios'
import { useRouter } from 'vue-router'

interface DbStats {
  words_total: number
  words_unique: number
  comics_fetched: number
  comics_total: number
}

type UpdateStatus = 'idle' | 'running'

const router = useRouter()
const message = useMessage()

// Состояния
const dbStats = ref<DbStats>({
  words_total: 0,
  words_unique: 0,
  comics_fetched: 0,
  comics_total: 0,
})

const updateStatus = ref<UpdateStatus>('idle')
const isLoading = ref(false)
const isUpdating = ref(false)
const isDropping = ref(false)

// Вычисляемые свойства
const updateStatusConfig = computed(() => {
  const isRunning = updateStatus.value === 'running'
  return {
    tagType: isRunning ? 'warning' : 'success',
    alertType: isRunning ? 'warning' : 'info',
    title: isRunning ? 'DB is updating..' : 'DB is ready',
    message: isRunning ? 'DB is currently updating. It may take some time.' : 'DB is idle.',
    disabled: isRunning || isUpdating.value,
  }
})

// Методы
const fetchDbStats = async () => {
  try {
    const { data } = await apiClient.get<DbStats>('/api/db/stats')
    dbStats.value = data
  } catch (error: any) {
    handleError(error, 'Failed to load DB stats')
  }
}

const fetchUpdateStatus = async () => {
  try {
    const { data } = await apiClient.get<{ status: UpdateStatus }>('/api/db/status')
    updateStatus.value = data.status
  } catch (error: any) {
    handleError(error, 'Failed to check update status')
  }
}

const handleUpdate = async () => {
  isUpdating.value = true
  updateStatus.value = 'running'

  try {
    await apiClient.post('/api/db/update', null, { timeout: 300000 })

    // После успешного старта обновления, периодически проверяем статус
    await waitForUpdateCompletion()
  } catch (error: any) {
    if (error.response?.status === 401) {
      router.push('login')
    } else {
      message.error('Failed to start update')
      // console.error('Update error:', error)
      updateStatus.value = 'idle'
    }
  } finally {
    isUpdating.value = false
  }
}

const waitForUpdateCompletion = async () => {
  const maxAttempts = 3
  const interval = 50

  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    await new Promise((resolve) => setTimeout(resolve, interval))

    try {
      await fetchUpdateStatus()
      // Обновлене завершено, обновляем статистику
      await fetchDbStats()
      message.success('Update completed successfully')
      return
    } catch (error) {
      console.error('Error checking update status:', error)
    }
  }

  message.warning('Update check timeout. Status may not be accurate.')
}

const handleDrop = async () => {
  isDropping.value = true
  try {
    await apiClient.delete('/api/db')
    message.success('All comics deleted successfully')
    await fetchDbStats()
  } catch (error: any) {
    handleError(error, 'Failed to drop comics')
  } finally {
    isDropping.value = false
  }
}

const handleError = (error: any, defaultMessage: string) => {
  if (error.response?.status === 401) {
    // message.error('Token is expired or invalid')
    router.push('/login')
  } else {
    message.error(defaultMessage)
  }
  // console.error(error)
}

const refreshAll = async () => {
  isLoading.value = true
  try {
    await Promise.all([fetchDbStats(), fetchUpdateStatus()])
    // message.info("Stats & Update refreshed!")
  } finally {
    isLoading.value = false
  }
}

// Хуки жизненного цикла
onMounted(() => {
  refreshAll()
})

// Следим за изменением статуса обновления
watch(updateStatus, (newStatus) => {
  if (newStatus === 'idle') {
    fetchDbStats()
  }
})
</script>

<template>
  <div class="admin-container">
    <n-card class="admin-card">
      <!-- HEADER -->
      <div class="header">
        <div class="logo" @click="router.push('/')">
          <img src="/logo.png" alt="xkcd logo" />
        </div>
        <h1>Admin Panel</h1>
      </div>

      <!-- DATABASE STATS -->
      <n-card size="small" class="stats-section" bordered>
        <n-space vertical size="large">
          <h2>Database Stats</h2>
          <n-grid cols="1 s:2 m:4" :x-gap="16" :y-gap="16">
            <n-gi>
              <n-statistic label="Words Total" :value="dbStats.words_total" />
            </n-gi>
            <n-gi>
              <n-statistic label="Words Unique" :value="dbStats.words_unique" />
            </n-gi>
            <n-gi>
              <n-statistic
                label="Comics in DB"
                :value="dbStats.comics_fetched + ' / ' + dbStats.comics_total"
              />
            </n-gi>
          </n-grid>
        </n-space>
      </n-card>

      <!-- DATABASE ACTIONS -->
      <n-card size="small" class="actions-section" bordered>
        <n-space vertical size="large">
          <div class="status-row">
            <n-space align="center" justify="space-between" style="width: 100%">
              <n-space align="center">
                <h2 class="status-label">Update Status:</h2>
                <n-tag :type="updateStatus === 'idle' ? 'info' : 'warning'">
                  {{ updateStatus }}
                </n-tag>
              </n-space>

              <n-button text type="primary" @click="refreshAll" :loading="isLoading">
                Refresh
              </n-button>
            </n-space>
          </div>

          <div class="actions-buttons">
            <n-button
              type="primary"
              size="large"
              @click="handleUpdate"
              :loading="updateStatus === 'idle' ? false : true"
              :disabled="updateStatusConfig.disabled"
              class="update-button"
            >
              Update DB
            </n-button>

            <n-popconfirm
              :positive-text="'Drop'"
              :negative-text="'Cancel'"
              @positive-click="handleDrop"
              show-icon
              class="delete-button"
            >
              <template #trigger>
                <n-button
                  type="error"
                  size="large"
                  :loading="isDropping"
                  :disabled="updateStatusConfig.disabled"
                  class="delete-button"
                >
                  Drop DB
                </n-button>
              </template>
              Warning!<br>
              This action will delete all data from the database.<br>
              Continue?
            </n-popconfirm>
          </div>
        </n-space>
      </n-card>
    </n-card>
  </div>
</template>

<style scoped>
.admin-container {
  /* min-height: 100vh; */
  display: flex;
  justify-content: center;
  align-items: flex-start;
  /* background: var(--color-background); */
  padding: 20px 4px 40px;
  flex: 1;
}

.admin-card {
  /* background: var(--surface-bg); */
  border-radius: var(--surface-radius);
  border: 1.5px solid var(--surface-border);
  width: 100%;
  max-width: 900px;
  width: 100%;
}

.header {
  display: flex;
  align-items: center;
  gap: 16px;
  border-bottom: 1px solid var(--divider-color);
  margin-bottom: 16px;
}

.logo img {
  max-width: 140px;
  cursor: pointer;
}
.logo img {
  width: 100%;
  height: auto;
  object-fit: contain;
}

.logo:hover {
  opacity: 0.8;
}

.status-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-label {
  font-weight: 600;
}

.n-tag {
  text-transform: capitalize;
}

.actions-buttons {
  display: flex;
  gap: 16px;
  width: 100%;
  margin-top: 12px;
  flex-wrap: wrap;
}

.update-button {
  flex: 1 1 45%;
  width: 100%;
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
  /* min-width: 300px; */
}

.delete-button {
  flex: 1 1 45%;
  width: 100%;
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
  /* min-width: 300px; */
}
</style>
