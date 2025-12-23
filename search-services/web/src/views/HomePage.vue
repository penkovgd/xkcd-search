<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { NTabs, NTabPane, NInput } from 'naive-ui'
import CategoriesPanelComponent from '@/components/CategoriesPanelComponent.vue'

const router = useRouter()
const searchType = ref('default')
const searchValue = ref('')

const handleSearch = () => {
  const phrase = searchValue.value.trim()
  if (!phrase) return

  router.push({
    name: 'search',
    query: {
      phrase,
      mode: searchType.value === 'index' ? 'index' : 'default',
    },
  })
}

const goToHome = () => {
  router.push('/')
}
</script>

<template>
  <div class="home-container">
    <div class="layout-wrapper">
      <categories-panel-component class="categories-panel"/>
      <div class="content-card surface">
        <!-- Logo -->
        <div class="logo-container" @click="goToHome">
          <img class="logo-image" src="/logo.png" alt="xkcd logo" />
        </div>
  
        <!-- Tabs -->
        <div class="tabs-container">
          <n-tabs v-model:value="searchType" type="line" size="large" justify-content="space-evenly">
            <n-tab-pane name="default" tab="Default search" style="font-weight: 600" />
            <n-tab-pane name="index" tab="Index search" />
          </n-tabs>
        </div>
  
        <!-- Search Input -->
        <div class="search-container">
          <n-input
            v-model:value="searchValue"
            placeholder="Search for comics.."
            size="large"
            @keyup.enter="handleSearch"
          >
          </n-input>
          <n-button color="#6e7b91" class="search-button" @click="handleSearch">Search!</n-button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.layout-wrapper {
  display: flex;
  width: 100%;
  justify-content: center;
  gap: 8px;
  max-height: 480px;
}

.categories-panel {
  width: 300px;
}

.home-container {
  /* min-height: 100vh; */
  display: flex;
  align-items: center;
  justify-content: center;
  /* background-color: var(--color-background); */
  padding: 4px;
  flex: 1;
}

.content-card {
  /* background: var(s--surface-bg); */
  border-radius: var(--surface-radius);
  border: 1.5px solid var(--surface-border);
  padding: 60px 40px;
  width: 100%;
  max-width: 600px;
}

@media (max-width: 768px) {
  .layout-wrapper {
    flex-direction: column-reverse;
    align-items: center;
    max-height: fit-content;
  }
  .categories-panel {
    width: 600px;
  }

  .content-card {
    padding: 30px 20px;
  }
}

@media (max-width: 480px) {
  .content-card {
    padding: 20px 15px;
  }
}

.logo-container {
  display: flex;
  justify-content: center;
  margin: 0 auto;
  margin-bottom: 40px;
  width: 100%;
  overflow: hidden;
  cursor: pointer;
  max-width: 300px;
}

.logo-image {
  max-width: 100%;
  width: auto;
  max-height: 300px;
  height: auto;
  object-fit: contain;
  transition: opacity 0.2s;
}

.logo-container:hover .logo-image {
  opacity: 0.8;
}

.tabs-container {
  /* margin-bottom: 30spx; */
  display: flex;
  justify-content: center;
}

.search-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
}

/* .search-button:hover {
  background-color: #f5f5f5;
}

.search-button:active {
  background-color: #e8e8e8;
} */
</style>
