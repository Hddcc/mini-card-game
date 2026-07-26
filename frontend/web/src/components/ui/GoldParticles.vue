<script setup lang="ts">
/** 金色上升粒子背景（迁移自旧 mini_1 登录页 canvas 粒子系统，30 个粒子）。 */
import { onBeforeUnmount, onMounted, ref } from 'vue'

interface Particle {
  x: number
  y: number
  size: number
  speedY: number
  opacity: number
}

const canvasRef = ref<HTMLCanvasElement | null>(null)

let particles: Particle[] = []
let rafId = 0

function initParticles(): void {
  const canvas = canvasRef.value
  if (!canvas) return
  canvas.width = window.innerWidth
  canvas.height = window.innerHeight
  particles = []
  for (let i = 0; i < 30; i++) {
    particles.push({
      x: Math.random() * canvas.width,
      y: Math.random() * canvas.height,
      size: Math.random() * 2 + 1,
      speedY: Math.random() * -0.5 - 0.2,
      opacity: Math.random() * 0.5 + 0.1,
    })
  }
}

function animate(): void {
  const canvas = canvasRef.value
  const ctx = canvas?.getContext('2d')
  if (!canvas || !ctx) return
  ctx.clearRect(0, 0, canvas.width, canvas.height)
  for (const particle of particles) {
    particle.y += particle.speedY
    if (particle.y < 0) particle.y = canvas.height
    ctx.fillStyle = `rgba(255, 215, 0, ${particle.opacity})`
    ctx.beginPath()
    ctx.arc(particle.x, particle.y, particle.size, 0, Math.PI * 2)
    ctx.fill()
  }
  rafId = requestAnimationFrame(animate)
}

onMounted(() => {
  initParticles()
  animate()
  window.addEventListener('resize', initParticles)
})

onBeforeUnmount(() => {
  cancelAnimationFrame(rafId)
  window.removeEventListener('resize', initParticles)
})
</script>

<template>
  <canvas ref="canvasRef" class="pointer-events-none fixed inset-0 z-[5]" aria-hidden="true"></canvas>
</template>
