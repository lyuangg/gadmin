# Gadmin

## 介绍

基于 Gin + GORM + MySQL 的后台管理系统，前端使用 Go 模板 + Vue 3 + Element Plus，支持 JWT 认证、RBAC 权限、操作日志与定时任务。

## 功能

- **认证**：用户名密码登录、图片验证码、JWT Token（支持退出失效）
- **权限**：角色-权限 RBAC、超级管理员、路由级权限、菜单按权限展示
- **用户管理**：用户 CRUD、角色分配、启用/禁用、重置密码
- **角色管理**：角色 CRUD、权限分配
- **权限管理**：权限 CRUD、从路由自动扫描导入
- **操作日志**：记录 PUT/DELETE/POST 请求与响应，支持按时间/用户/方法/路径筛选与分页
- **定时任务**：每天凌晨清理操作日志，保留最近 N 条（可配置）
- **个人中心**：修改密码、更换头像

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.23、Gin、GORM、MySQL、JWT、robfig/cron |
| 前端 | Go 模板、Vue 3、Element Plus、Axios |
| 部署 | Docker、Docker Compose |

## 项目结构

```
gadmin/
├── main.go                 # 入口：加载配置、初始化应用、路由、定时任务、启动 HTTP
├── app/                    # 应用上下文（DB、Logger、Services）
├── config/                 # 配置定义与加载（YAML + 环境变量）
├── database/               # 数据库初始化与迁移
├── models/                 # 数据模型（User、Role、Permission、OperationLog）
├── services/               # 业务逻辑（Auth、User、Role、Permission、OperationLog）
├── controllers/            # HTTP 控制器
├── middleware/             # JWT 认证、权限校验、操作日志记录、路由扫描导入权限
├── routes/                 # 路由注册与模板渲染
├── tasks/                  # 定时任务（操作日志每日清理，基于 robfig/cron）
├── utils/                  # JWT、验证码、统一响应等工具
├── templates/              # HTML 模板（布局、登录、管理页、分页组件）
├── static/                 # 前端静态资源（JS/CSS/Element Plus/Vue/Axios）
├── config.yml.example      # 配置示例
├── Dockerfile
└── docker-compose.yml
```

## 快速开始

### 使用 Docker Compose（推荐）

```bash
docker-compose up -d
```

- **应用访问**：<http://localhost:8899>
- **MySQL**：宿主机端口 3308 映射容器 3306

### 默认账号

| 用户名 | 密码     | 角色       |
|--------|----------|------------|
| admin  | admin123 | 超级管理员 |

## 本地开发

### 环境要求

- Go 1.23+
- MySQL 8.0+

### 配置

复制配置示例并按需修改：

```bash
cp config.yml.example config.yml
```

主要配置项见下表，未在 YAML 中配置的项可从环境变量读取（如 `DB_HOST`、`PORT`、`OPERATION_LOG_RETAIN_COUNT` 等），详见 `config/config.go` 与 `config.yml.example`。

| 配置项 | 说明 | 默认/示例 |
|--------|------|------------|
| db_host / db_port / db_user / db_password / db_name | 数据库连接 | localhost, 3308, root, ***, gadmin |
| db_table_prefix | 表前缀 | 空或 `gadmin_` |
| jwt_secret | JWT 密钥 | 生产环境务必修改 |
| port | 服务端口 | 8080 |
| gin_mode | debug / release / test | release |
| log_type / log_level / log_output | 日志格式、级别、输出 | text, info, 空=标准输出 |
| operation_log_retain_count | 操作日志保留条数（每日凌晨清理） | 10000 |

### 运行

```bash
# 使用默认或环境变量
go run main.go

# 指定配置文件
go run main.go -c ./config.yml
```

## 部署

### Docker 构建与运行

```bash
docker build -t gadmin .

docker run -d -p 8080:8080 \
  -e DB_HOST=mysql \
  -e DB_PORT=3306 \
  -e DB_USER=root \
  -e DB_PASSWORD=xxx \
  -e DB_NAME=gadmin \
  -e JWT_SECRET=your-secret \
  gadmin
```

### Docker Compose

```bash
docker-compose up -d
```

应用在容器内监听 8080，compose 中映射为宿主机 8899。

## 许可证

MIT
