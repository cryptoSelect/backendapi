# CryptoSelect Backend

加密货币选币分析 API 服务，基于 Go + Gin，提供标的与资金流向等接口。

## 技术栈

- **Go** 1.25+
- **Gin** - HTTP 框架
- **GORM** + **PostgreSQL** - 数据库（通过 `cryptoSelect/public` 公共库）
- **CORS** - 跨域支持

## 项目结构

```
backend/
├── main/           # 入口与路由注册
├── api/
│   ├── symbol/     # 标的相关接口
│   └── funds/      # 资金流向接口
├── config/         # 配置加载与示例
├── utils/logger/   # 日志
└── Dockerfile
```

## 环境要求

- Go 1.25+
- PostgreSQL
- 依赖仓库：`github.com/cryptoSelect/public`（数据库等公共逻辑）

## 配置

1. 复制配置示例并填写实际值：

   ```bash
   cp config/config.example.json config/config.json
   ```

2. 编辑 `config/config.json`：

   | 字段 | 说明 |
   |------|------|
   | `Mode` | 运行模式：`prod` / `debug` |
   | `Database.Host` | 数据库主机 |
   | `Database.Port` | 数据库端口 |
   | `Database.User` | 数据库用户 |
   | `Database.Password` | 数据库密码 |
   | `Database.DBName` | 数据库名 |
   | `Database.SSLMode` | 如 `disable` |
   | `Page.PageSize` | 默认分页大小 |

> 请勿将 `config/config.json` 提交到版本库（已加入 `.gitignore`）。

## 本地运行

```bash
# 安装依赖
go mod download

# 启动服务（默认 :8080）
go run main/main.go
```

## Docker 运行

```bash
# 构建镜像
docker build -t cryptoselect-backend .

# 运行前需将 config/config.json 放到当前目录或挂载进容器
docker run -p 8080:8080 -v $(pwd)/config:/app/config cryptoselect-backend
```

镜像内默认工作目录为 `/app`，入口为 `./backend`，端口 **8080**。

## API 概览

基础路径：`/api`，默认支持 CORS。

| 模块 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 标的 | GET | `/api/symbol/` | 查询标的列表 |
| 资金流向 | GET | `/api/funds/` | 查询资金流向数据 |

具体请求参数与响应格式见各 handler 实现。

## License

见 [LICENSE](./LICENSE)。
