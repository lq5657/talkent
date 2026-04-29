<script setup lang="ts">
import { computed } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import hljs from 'highlight.js'

marked.use({
  renderer: {
    code({ text, lang }: { text: string; lang?: string }) {
      const language = lang || ''
      if (language && hljs.getLanguage(language)) {
        return `<pre><code class="hljs language-${language}">${hljs.highlight(text, { language }).value}</code></pre>`
      }
      return `<pre><code class="hljs">${hljs.highlightAuto(text).value}</code></pre>`
    },
  },
})

const props = defineProps<{
  content: string
}>()

const rendered = computed(() => {
  const raw = marked.parse(props.content) as string
  return DOMPurify.sanitize(raw)
})
</script>

<template>
  <div class="prose prose-sm max-w-none" v-html="rendered" />
</template>
