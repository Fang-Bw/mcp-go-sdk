# Document Reader MCP Server

这是一个支持读取多种文档格式的 MCP 服务器，可以读取 PDF、Word、文本和 Markdown 文件。

## 功能

- **read_file**: 读取本地文件内容
  - 支持 `.txt` - 纯文本文件
  - 支持 `.md` - Markdown 文件
  - 支持 `.pdf` - PDF 文档（提取文本内容）
  - 支持 `.docx` - Word 文档（.docx 格式）

## 已安装的库

- `github.com/ledongthuc/pdf` - PDF 文本提取
- `github.com/nguyenthenguyen/docx` - Word 文档读取

## 编译

```bash
cd examples/server/document-reader
go build -o mcp-document-reader.exe main.go
```

## 在 Cursor 中配置

在 `C:\Users\{用户名}\.cursor\mcp.json` 文件中添加：

```json
{
  "mcpServers": {
    "document-reader": {
      "command": "D:\\golandProjects\\thirdparty\\mcp-go-sdk\\examples\\server\\document-reader\\mcp-document-reader.exe",
      "args": []
    }
  }
}
```

配置完成后，重启 Cursor 使配置生效。

## 使用示例

配置完成后，可以在 Cursor 中通过 AI 助手调用：

1. **读取文本文件**: "读取 D:\documents\readme.txt 文件"
2. **读取 PDF**: "读取 D:\documents\report.pdf 文件内容"
3. **读取 Word**: "读取 D:\documents\document.docx 文件"

## 注意事项

1. **文件路径**: 支持绝对路径和相对路径
   - 绝对路径: `D:\documents\file.pdf`
   - 相对路径: `./documents/file.pdf`（相对于服务器运行目录）

2. **文件格式限制**:
   - 不支持旧的 `.doc` 格式（仅支持 `.docx`）
   - PDF 文件会提取所有页面的文本内容
   - Word 文档会提取所有段落内容

3. **文件大小**: 对于非常大的文件，处理可能需要一些时间

4. **权限**: 确保服务器有权限读取指定的文件

## 技术细节

- PDF 解析使用 `github.com/ledongthuc/pdf` 库，逐页提取文本
- Word 文档解析使用 `github.com/nguyenthenguyen/docx` 库，提取所有段落内容
- 文本和 Markdown 文件直接读取文件内容

