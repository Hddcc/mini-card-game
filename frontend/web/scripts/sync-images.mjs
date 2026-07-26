/**
 * 把旧前端的游戏美术资源同步到 public/，使生产构建产物在
 * /static/assets/images/ 下提供与后端 seed（internal/model/db.go）
 * 一致的绝对图片路径。仅同步 images/ 子目录（cards/portraits 的
 * svg 为未引用的占位资源，不迁移）。
 *
 * 注意：不用 fs.cpSync —— 该 API 在本机（Node 22 + 含中文的项目路径）
 * 会原生崩溃（无异常直接退出），改用 readdir + copyFile 手工递归。
 */
import { copyFileSync, existsSync, mkdirSync, readdirSync, statSync } from 'node:fs'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))
const source = resolve(here, '../../stitch/assets/images')
const target = resolve(here, '../public/assets/images')

if (!existsSync(source)) {
  console.error(`[sync-images] 源目录不存在: ${source}`)
  process.exit(1)
}

let copied = 0

function copyDir(from, to) {
  mkdirSync(to, { recursive: true })
  for (const name of readdirSync(from)) {
    const fromPath = join(from, name)
    const toPath = join(to, name)
    if (statSync(fromPath).isDirectory()) {
      copyDir(fromPath, toPath)
    } else {
      copyFileSync(fromPath, toPath)
      copied += 1
    }
  }
}

copyDir(source, target)
console.log(`[sync-images] 已同步 ${copied} 个文件 -> ${target}`)
