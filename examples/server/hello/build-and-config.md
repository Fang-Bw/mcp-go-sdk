# 编译及配置步骤

编译
```
go build -o mcp-hello-server.exe main.go
```

配置（Cursor的mcp.json中）
```
{
  "mcpServers": {
    "greeter": {
      "command": "D:\\golandProjects\\thirdparty\\mcp-go-sdk\\examples\\server\\hello\\mcp-hello-server.exe",
      "args": []
    }
  }
}
```