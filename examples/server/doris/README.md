# Doris MCP Server

这是一个连接 Apache Doris 数据库的 MCP 服务器，提供了查询、列出表和描述表结构等功能。

## 功能

- **query**: 执行 SQL 查询并返回结果
- **list_tables**: 列出指定数据库中的所有表
- **describe_table**: 获取表的结构和列信息

## 编译

```bash
cd examples/server/doris
go build -o mcp-doris-server.exe main.go
```

## 配置

### 方式 1: 命令行参数

```bash
mcp-doris-server.exe -host localhost -port 9030 -user root -password your_password -database your_database
```

### 方式 2: 环境变量

可以通过环境变量设置密码和数据库：

```bash
set DORIS_PASSWORD=your_password
set DORIS_DATABASE=your_database
mcp-doris-server.exe -host localhost -port 9030 -user root
```

### 方式 3: 混合使用

命令行参数和环境变量可以混合使用，命令行参数优先级更高。

## 在 Cursor 中配置

在 `C:\Users\{用户名}\.cursor\mcp.json` 文件中添加：

```json
{
  "mcpServers": {
    "doris": {
      "command": "D:\\golandProjects\\thirdparty\\mcp-go-sdk\\examples\\server\\doris\\mcp-doris-server.exe",
      "args": [
        "-host", "localhost",
        "-port", "9030",
        "-user", "root",
        "-password", "your_password",
        "-database", "your_database"
      ]
    }
  }
}
```

**注意**: 如果不想在配置文件中暴露密码，可以使用环境变量：

```json
{
  "mcpServers": {
    "doris": {
      "command": "D:\\golandProjects\\thirdparty\\mcp-go-sdk\\examples\\server\\doris\\mcp-doris-server.exe",
      "args": [
        "-host", "localhost",
        "-port", "9030",
        "-user", "root",
        "-database", "your_database"
      ],
      "env": {
        "DORIS_PASSWORD": "your_password"
      }
    }
  }
}
```

配置完成后，重启 Cursor 使配置生效。

## 使用示例

配置完成后，可以在 Cursor 中通过 AI 助手调用这些工具：

1. **执行查询**: "查询用户表中的前10条记录"
2. **列出表**: "列出当前数据库中的所有表"
3. **描述表**: "显示 users 表的结构"

## 参数说明

- `-host`: Doris 服务器地址（默认: localhost）
- `-port`: Doris 服务器端口（默认: 9030）
- `-user`: 数据库用户名（默认: root）
- `-password`: 数据库密码（可通过环境变量 DORIS_PASSWORD 设置）
- `-database`: 默认数据库名（可通过环境变量 DORIS_DATABASE 设置）

## 注意事项

1. Doris 使用 MySQL 协议，端口通常是 9030（FE）或 9060（查询端口）
2. 确保 Doris 服务器允许来自客户端的连接
3. 密码建议使用环境变量或配置文件，避免在命令行中暴露
4. 服务器会维护连接池，提高查询性能

