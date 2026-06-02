# Docker 部署说明

这套部署会同时启动应用、MySQL、Redis。服务器只需要安装 Docker 和 Docker Compose。

## 0. 服务器安装 Docker

如果腾讯云服务器是新机器，可以先执行：

```bash
curl -fsSL https://get.docker.com | sudo sh
sudo systemctl enable --now docker
docker compose version
```

如果提示当前用户没有 Docker 权限，先使用 `sudo docker compose ...`，或执行：

```bash
sudo usermod -aG docker $USER
```

然后重新登录服务器。

## 1. 上传项目

把 `mini-card-game` 整个目录上传到服务器，例如：

```bash
scp -r mini-card-game root@你的服务器IP:/opt/mini-card-game
```

## 2. 启动

```bash
cd /opt/mini-card-game
cp .env.example .env
docker compose up -d --build
```

也可以使用项目里的快捷脚本：

```bash
bash deploy/tencentcloud-quickstart.sh
```

访问：

```text
http://你的服务器IP:5290
```

腾讯云安全组需要放行 TCP `5290` 端口。

## 常用命令

```bash
docker compose ps
docker compose logs -f app
docker compose restart app
docker compose pull
docker compose down
```

保留数据并升级：

```bash
docker compose up -d --build
```

清空数据库和 Redis 数据：

```bash
docker compose down -v
```

## 配置

可在 `.env` 中修改：

- `APP_PORT`：对外访问端口，默认 `5290`
- `MYSQL_ROOT_PASSWORD`：MySQL root 密码，建议部署前修改
- `JWT_SECRET`：登录令牌密钥，建议部署前改成长随机字符串

MySQL 和 Redis 默认只在 Docker 内部网络暴露，公网只开放应用端口。
