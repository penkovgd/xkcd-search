<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NSpin } from 'naive-ui'
import apiClient from '@/utils/axios'

type CategoryWithCount = {
  category: string
  count: number
}

const router = useRouter()

const loading = ref(false)
const errorMessage = ref<string | null>(null)
const categories = ref<CategoryWithCount[]>([])

const totalCount = computed(() => categories.value.reduce((s, c) => s + (c.count || 0), 0))

const fetchCategories = async () => {
  loading.value = true
  errorMessage.value = null
  try {
    const { data } = await apiClient.get<{ categories: CategoryWithCount[] }>('/api/categories')
    categories.value = data.categories || []
  } catch (err) {
    console.error('Failed to load categories', err)
    errorMessage.value = 'Failed to load categories'
    categories.value = []
  } finally {
    loading.value = false
  }
}

const goToCategory = (cat: string) => {
  router.push({ name: 'category', params: { category: cat } }).catch(() => {})
}

onMounted(() => {
  fetchCategories()
})
</script>

<template>
    <div class="surface categories-card">
      <!-- Categories list -->
      <div class="form-section">
        <h2 class="form-title">Categories</h2>

        <div class="categories-body">
          <n-spin :show="loading">
            <div v-if="errorMessage" class="error-text">{{ errorMessage }}</div>

            <ul class="categories-list" v-else>
              <li class="category-item">
                <a class="category-link" href="#" @click.prevent="goToCategory('')">
                  <span class="category-name">All</span>
                  <span class="category-count">({{ totalCount }})</span>
                </a>
              </li>

              <li
                v-for="c in categories"
                :key="c.category"
                class="category-item"
              >
                <a
                  class="category-link"
                  href="#"
                  @click.prevent="goToCategory(c.category)"
                >
                  <span class="category-name">{{ c.category }}</span>
                  <span class="category-count">({{ c.count }})</span>
                </a>
              </li>
            </ul>
          </n-spin>
        </div>
      </div>
    </div>
</template>

<style scoped>

.categories-card {
  display: flex;
  flex-direction: column;
  padding: 24px;
  max-width: 100%;
  max-height: 780.6px;
}

.form-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.form-title {
  font-size: 22px;
  font-weight: 600;
  color: var(--text-dark);
  margin: 0 0 20px 0;
  /* text-align: center; */
}

.categories-body {
  flex: 1;
  max-height: 100%;
  overflow-y: auto;
  padding-right: 8px;
  margin-bottom: 20px;
}

.categories-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.category-item {
  display: flex;
  /* align-items: center; */
}

.category-link {
  display: flex;
  justify-content: space-between;
  /* align-items: center; */
  width: 100%;
  /* text-decoration: none; */
  padding: 4px 0px;
  /* border-radius: 8px; */
  /* border: 1px solid transparent; */
  /* transition: background 0.2s ease, border-color 0.2s ease; */
  color: var(--text-dark);
  /* background: var(--surface-bg); */
}

.category-link:hover {
  background: rgba(150, 168, 200, 0.2);
  /* border-color: var(--surface-border); */
  /* text-decoration: none; */
}

.category-name {
  font-weight: 500;
  font-size: 14px;
}

.category-count {
  color: var(--text-gray);
  font-size: 14px;
  margin-left: 12px;
  text-decoration: none;
}

.actions-row {
  display: flex;
  justify-content: center;
  padding-top: 16px;
  border-top: 1px solid var(--surface-border);
}

.error-text {
  color: #ff4d4f;
  text-align: center;
  padding: 20px;
}

/* Убираем ненужные стили */
.panel-container {
  display: none;
}

.login-card {
  display: none;
}

/* Адаптивные стили */
@media (max-width: 768px) {
  .categories-card {
    /* min-height: fit-content; */
    padding: 24px;
    max-height: fit-content;
  }
  
  .form-title {
    font-size: 20px;
    margin-bottom: 16px;
  }
  
  .category-link {
    padding: 6px 0px;
  }
}

.surface {
  background-color: var(--surface-bg);
  border-radius: var(--surface-radius);
  border: 1.5px solid var(--surface-border);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}
</style>