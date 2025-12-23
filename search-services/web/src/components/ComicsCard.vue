<script setup lang="ts">
import {
  NCard,
  NSpin,
  NSkeleton,
  NEllipsis,
  NTag,
  NText,
} from 'naive-ui'

// входные данные
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const props = defineProps<{
  comic: {
    ID: number
    URL: string
    Title: string
    Date: string
    Category: string
  }
  imageLoading: Record<number, boolean>
  setImageRef: (el: HTMLImageElement | null, comicId: number) => void
  handleImageLoad: (comicId: number) => void
  handleImageError: (event: Event) => void
  openComic: (comicId: number) => void
}>()

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
</script>

<template>
  <n-card
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
        :ref="(el) => setImageRef(el as HTMLImageElement, comic.ID)"
        v-show="!imageLoading[comic.ID]"
      />
    </template>

    <!-- Comic title -->
    <template #header>
      <div class="comic-title">
        <n-ellipsis line-clamp="1">
          {{ comic.Title }}
        </n-ellipsis>
      </div>
    </template>

    <!-- Comic metadata -->
    <template #footer>
      <n-ellipsis><n-text>{{ comic.Category }}</n-text></n-ellipsis>
      <div class="comic-meta">
        <div class="comic-date">
          <n-text depth="3">
            {{ formatDate(comic.Date) }}
          </n-text>
        </div>
        <n-tag size="small" type="info" class="id-tag">
          #{{ comic.ID }}
        </n-tag>
      </div>
    </template>
  </n-card>
</template>

<style scoped>
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
  max-height: 24px;
  font-weight: 600;
}
.comic-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.comic-date {
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
</style>
