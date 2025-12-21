<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NInput, NButton, NSpin, NCard, NEllipsis, NTag, NText, type SelectOption } from 'naive-ui'
import apiClient from '@/utils/axios'

type Comic = {
  ID: number
  URL: string
  Title: string
  Date: string
}
type SearchResponse = {
  comics: Comic[]
  total: number
}

const route = useRoute()
const router = useRouter()

const searchQuery = ref<string>((route.query.phrase as string) || '')
const searchMode = ref<string>((route.query.mode as string) || '')
const selectedMode = ref<string>((route.query.mode as string) || '')

const modeOptions: SelectOption[] = [
  {
    label: 'Default',
    value: 'default',
  },
  {
    label: 'Index',
    value: 'index',
  },
]

const apiEndpoint = computed(() => {
  return searchMode.value === 'index' ? '/api/isearch' : '/api/search'
})

const displayPhrase = ref<string>((route.query.phrase as string) || '')
const loading = ref(false)
const imageLoading = ref<Record<number, boolean>>({})
const errorMessage = ref<string | null>(null)
const comics = ref<Comic[]>([])

const totalFound = computed(() => comics.value.length)

const currentPage = ref(1)
const pageSize = 15

const paginatedComics = computed(() => {
  const startIndex = (currentPage.value - 1) * pageSize
  const endIndex = startIndex + pageSize
  return comics.value.slice(startIndex, endIndex)
})

const fetchResults = async () => {
  const phrase = searchQuery.value.trim()
  if (!phrase) {
    comics.value = []
    return
  }

  loading.value = true
  errorMessage.value = null

  try {
    const { data } = await apiClient.get<SearchResponse>(apiEndpoint.value, {
      params: {
        phrase: phrase,
        limit: 5000,
      },
    })

    comics.value = data.comics
    currentPage.value = 1

    data.comics.forEach((comic) => {
      imageLoading.value[comic.ID] = true
    })
    displayPhrase.value = phrase
    selectedMode.value = searchMode.value
  } catch {
    errorMessage.value = 'Failed to load results. Please try again.'
  } finally {
    loading.value = false
  }
}

const handleImageLoad = (comicId: number) => {
  imageLoading.value[comicId] = false
}

const submitSearch = () => {
  const phrase = searchQuery.value.trim()
  if (!phrase) return

  router.push({
    name: 'search',
    query: {
      phrase,
      mode: selectedMode.value,
    },
  })
}

const handleImageError = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.src =
    'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="300" height="400" viewBox="0 0 300 400"><rect width="300" height="400" fill="%23f5f5f5"/><text x="150" y="200" font-family="Arial" font-size="16" text-anchor="middle" fill="%23999">Image not available</text></svg>'
  img.onerror = null
}

const formatDate = (dateString: string) => {
  try {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  } catch {
    return 'Invalid date'
  }
}

const openComic = (comicId: number) => {
  const url = `https://xkcd.com/${comicId}/`
  window.open(url, '_blank', 'noopener,noreferrer')
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  // Прокрутка к верху контейнера результатов
  document.querySelector('body')?.scrollIntoView({ behavior: 'smooth' })
}

watch(
  () => route.query,
  (newQuery) => {
    searchQuery.value = (newQuery.phrase as string) || ''
    searchMode.value = (newQuery.mode as string) || 'default'
    fetchResults()
  },
  { immediate: true },
)

onMounted(() => {
  fetchResults()
})
</script>

<template>
  <div class="search-container">
    <div class="results-card surface">
      <!-- Header with logo and search -->
      <div class="header">
        <div class="logo" @click="router.push('/')">
          <img src="/logo.png" alt="xkcd logo" />
        </div>
        <div class="search-box">
          <n-input-group>
            <n-input
              v-model:value="searchQuery"
              placeholder="Search for comics.."
              size="large"
              clearable
              @keyup.enter="submitSearch"
            ></n-input>
            <n-select
              v-model:value="selectedMode"
              :options="modeOptions"
              :style="{
                width: '20%',
                minWidth: '100px',
                border: '1.5px solid var(--surface-border)',
              }"
              size="large"
              placeholder="Search mode"
              class="mode-selector"
            />
            <n-button size="large" @click="submitSearch" class="search-button" color="#6e7b91"
              >Search</n-button
            >
          </n-input-group>
        </div>
      </div>

      <!-- Results title -->
      <div class="results-title">
        <h2>Results for search: "{{ displayPhrase || '...' }}"</h2>
        <h3 v-if="!loading && !errorMessage" style="color: var(--text-gray); font-weight: 100">
          {{ totalFound }} result(s) found
        </h3>
        <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>
      </div>

      <div class="content">
        <n-spin :show="loading">
          <div v-if="!loading && comics.length === 0 && !errorMessage" class="empty-state">
            No results yet. Try another phrase.
          </div>

          <div class="grid" v-else>
            <n-card
              v-for="comic in paginatedComics"
              :key="comic.ID"
              class="comic-card"
              hoverable
              @click="openComic(comic.ID)"
            >
              <!-- Comic cover -->
              <template #cover>
                <div v-if="imageLoading[comic.ID]" class="image-loading">
                  <n-spin size="small">
                    <n-skeleton height="220px" width="100%" :sharp="false" />
                  </n-spin>
                </div>

                <img
                  :src="comic.URL"
                  :alt="comic.Title"
                  class="comic-image"
                  :class="{ 'image-loaded': !imageLoading[comic.ID] }"
                  @error="handleImageError"
                  @load="handleImageLoad(comic.ID)"
                  v-show="!imageLoading[comic.ID]"
                />
              </template>

              <!-- Comic title -->
              <template #header>
                <div class="comic-title">
                  <n-ellipsis line-clamp="2">
                    {{ comic.Title }}
                  </n-ellipsis>
                </div>
              </template>

              <!-- Comic metadata -->
              <template #footer>
                <div class="comic-meta">
                  <div class="comic-date">
                    <n-text depth="3">
                      {{ formatDate(comic.Date) }}
                    </n-text>
                  </div>
                  <n-tag size="small" type="info" class="id-tag"> #{{ comic.ID }} </n-tag>
                </div>
              </template>
            </n-card>
          </div>
        </n-spin>
        <div class="pagination-container">
          <n-pagination
            simple
            v-model:page="currentPage"
            :page-size="pageSize"
            :item-count="totalFound"
            @update:page="handlePageChange"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.search-container {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  /* background: var(--color-background); */
  padding: 20px 4px 40px;
}

.results-card {
  /* background: var(--surface-bg); */
  border-radius: var(--surface-radius);
  border: 1.5px solid var(--surface-border);
  width: 100%;
  max-width: 900px;
  padding: 20px;
}

.header {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 20px;
  align-items: center;
  margin-bottom: 16px;
}

.logo {
  display: flex;
  align-items: center;
  cursor: pointer;
  max-width: 140px;
}

.logo img {
  width: 100%;
  height: auto;
  object-fit: contain;
}

.logo:hover {
  opacity: 0.8;
}

.search-box {
  width: 100%;
}

.results-title {
  margin: 20px 0 20px;
}

.results-title h3 {
  font-size: 18px;
  margin-bottom: 4px;
  color: var(--text-dark);
}

.results-title p {
  margin: 0;
  color: var(--text-gray);
  font-size: 14px;
}

.error-text {
  color: #d03050;
}

.content {
  width: 100%;
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 20px;
}

.comic-card {
  min-height: 250px;
  height: 100%;
  display: flex;
  flex-direction: column;
  border: 1.5px solid var(--surface-border);
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease;
}

.comic-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.comic-image {
  padding: 4px;
  width: 100%;
  height: 220px;
  object-fit: contain;
  border-radius: 4px 4px 0 0;
}

.comic-title {
  flex: 1;
  /* padding: 12px 0 8px; */
  /* max-height: 16px; */
  font-weight: 600;
}

.comic-meta {
  display: flex;
  justify-content: space-between;
  align-items: top;
  /* padding-top: 8px; */
  /* border-top: 1px solid var(--n-border-color); */
}

.comic-id {
  display: flex;
  align-items: center;
}

.comic-date {
  font-size: 0.85rem;
  color: var(--text-gray);
}

.id-tag {
  background: var(--color-primary);
  border-color: var(--surface-border);
  font-variant: small-caps;
  font-family: Lucida, Helvetica, sans-serif;
  border: 1px solid #333;
  font-weight: 600;
  text-decoration: none;
  border-radius: 3px;
  box-shadow: 0 0 5px 0 gray;
  color: white;
}

.empty-state {
  text-align: center;
  color: var(--text-gray);
  padding: 40px 0;
  font-size: 1.1rem;
}

@media (max-width: 768px) {
  .grid {
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 16px;
  }

  .comic-image {
    height: 150px;
  }
}

@media (max-width: 640px) {
  .results-card {
    padding: 16px;
  }

  .header {
    grid-template-columns: 1fr;
    justify-items: center;
  }

  .logo {
    justify-content: center;
    margin-bottom: 16px;
  }

  .logo img {
    max-width: 120px;
  }

  .grid {
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 12px;
  }

  .comic-image {
    height: 120px;
  }

  .comic-title {
    min-height: 50px;
    font-size: 0.9rem;
  }
}

.pagination-container {
  margin-top: 24px;
  display: flex;
  justify-content: center;
}

@media (max-width: 768px) {
  .pagination-container {
    margin-top: 24px;
    padding-top: 16px;
  }
}
</style>
