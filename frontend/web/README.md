# Mini 西游 Web 前端（Vue 3 SPA）

对旧版 `frontend/stitch/` 七个独立 HTML 页的统一重构：单页应用、统一设计系统（黑金西游风）、统一 API 层与命名规范，**后端 Go 代码零改动**。旧目录保留作为参考与设计母本，不再维护。

## 技术栈

Vue 3（Composition API + `<script setup>`）· Vite（`base: '/static/'`）· TypeScript · Pinia · Vue Router（history 模式）· Tailwind CSS v3.4

设计 token 准源：`frontend/stitch/mythic_journey_redux/DESIGN.md`（57 色 + 7 级字体，已全量落入 `tailwind.config.js`）。字体与图标（Epilogue / Hanken Grotesk / JetBrains Mono / Material Symbols）通过 npm 自托管，不依赖 Google Fonts CDN。

## 本地开发

```bash
# 1. 启动后端（任选其一）
#    a) docker compose up -d          # 整套（后端跑在 :5290）
#    b) 本地 go run ./cmd/server      # 需自备 MySQL 并在 .env 配置 MYSQL_DSN / JWT_SECRET

# 2. 启动前端 dev server
cd frontend/web
npm install
npm run dev        # http://localhost:5173
```

dev 下 Vite 把 `/api` 与 `/static` 都代理到 `http://localhost:5290`：后端下发的绝对图片路径（如战斗单位 `card_art`）无需任何处理即可加载。

## 生产构建与切换（后端零改动）

```bash
cd frontend/web
npm run build      # 产物在 dist/：含 index.html、mini_1/code.html（Go 回落副本）、assets/、assets/images/
```

然后在仓库根 `.env` 增加一行并重启后端：

```
FRONTEND_DIST=frontend/web/dist
```

访问 `http://localhost:5290/` 即为新前端。**回滚**：删除该行重启，即恢复旧 stitch 页面。

原理：构建产物以 `/static/` 为 base，全部静态资源由 Go 现有的 `r.Static("/static", …)` 托管；`/` 与 NoRoute 回落 `mini_1/code.html`（构建时从 `index.html` 复制，兼容 router.go 的确定性分支），因此 `/heroes` 等深链接刷新可直达。前端路由路径不能带 `/static` 前缀（会被 gin 的静态路由吞掉）。

Docker 部署切换（可选）：宿主机先 `npm run build`，在 Dockerfile 增加 `COPY frontend/web/dist /app/frontend/web/dist`，并把 docker-compose.yml 的 `FRONTEND_DIST` 改为 `/app/frontend/web/dist`。

## 目录结构

```
src/
  constants/    # 全部魔法值唯一出处（体力/抽卡费用/关卡 seed 表/品质映射/存储 key/错误文案映射）
  types/        # 后端 21 个接口的请求/响应类型（照抄 Go json tag）
  api/          # http.ts 统一客户端 + 每业务域一个薄模块 + normalizers（PascalCase 资产归一化）
  stores/       # Pinia：auth / assets（体力倒计时）/ player / hero / battle（战斗状态机）
  router/       # 路由 + 守卫（requiresAuth / guestOnly / 战斗锁导航）
  composables/  # useToast / useLocalAvatar
  layouts/      # MainLayout（侧栏 + 顶栏资产条，替代旧 app-shell.js / status-bar.js）
  components/   # ui（原子组件）/ layout / hero / battle / gacha / activity
  views/        # Login / Home / Stages / Heroes / Gacha / StarUp / Activity
scripts/sync-images.mjs  # 同步 stitch/assets/images → public/（prebuild 自动执行）
```

## 命名与代码规范

- 组件 SFC：PascalCase（`HeroCard.vue`）；composable：`useXxx`；store：`useXxxStore`
- 服务端响应类型 `XxxView`、请求体 `XxxPayload`；常量 `UPPER_SNAKE`
- api 模块函数动词开头：`fetchHeroes` / `saveTeam` / `startBattle`
- 所有后端英文错误串在 `constants/messages.ts` 集中翻译；活动页关键字规则也在该文件

## 与后端对接的关键约定（重构时核实自 Go 源码）

- 响应包 `{code, message, data}`，成功 `code === 0`（字段是 `message` 不是 `msg`）
- `GET /heroes` 的 data 是 `{heroes: []}` 包裹；`/team`、`/stages/progress`、`/tasks/daily` 等是裸数组
- `POST /heroes/star-up` 返回的 `assets` 是 PascalCase（Go model 无 json tag），由 `api/normalizers.ts` 归一化
- 后端 NoRoute 对未知路径返回 HTML 200 → `http.ts` 检测 content-type 兜底
- 关卡配置无接口，`constants/stages.ts` 硬编码自 `internal/model/db.go` seed，改 seed 需同步
- 旧 `/api/v1/stage/fight` 一键结算接口已弃用（旧前端也未使用），新前端只对接回合制三件套

## 已知取舍

- Go 对 `/static` 响应 `Cache-Control: no-store`（后端既有行为），静态资源不走浏览器缓存；Material Symbols 字体约 4MB，首屏偏重。不改后端的前提下无法优化；可换回 Google Fonts `<link>`（旧页写法）以利用 CDN 缓存
- 本机 Node 22 的 `fs.cpSync` 在含中文的项目路径下会原生崩溃，`scripts/sync-images.mjs` 已改用 readdir+copyFile 手工递归；同理项目锁定 Tailwind v3（纯 JS），勿升 v4（原生二进制）
- 本地头像存 `localStorage`（key 与旧版一致 `mini-xiyou-local-avatar`），跨账号共享、不上传服务器（旧版行为）
- 旧页的装饰性元素（关卡页"自动战斗"按钮、死资源 svg）未迁移；阵容页的品质筛选/搜索从装饰升级为真实功能
