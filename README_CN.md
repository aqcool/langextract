# LangExtract Go

[English](README.md) | 中文

一个基于 LLM 从非结构化文本中提取结构化信息的 Go 库。

## 概述

LangExtract Go 是 Google [LangExtract](https://github.com/google/langextract) 库的 Golang 实现。它使用大语言模型（LLM）根据用户定义的指令和示例从文本文档中提取结构化信息。

## 特性

- **结构化信息提取**：使用 LLM 从非结构化文本中提取结构化数据
- **精确源文本定位**：将每个提取结果映射到源文本的确切位置
- **OpenAI API 兼容**：支持标准 OpenAI API（包括兼容 OpenAI API 的服务）
- **长文档处理**：文本分块、并行处理、多轮提取以提升召回率
- **结构化输出**：基于 Schema 约束的 JSON/YAML 输出格式
- **文本对齐**：LCS（最长公共子序列）算法实现模糊文本匹配

## 安装

```bash
go get github.com/aqcool/langextract
```

## 快速开始

### 基础提取

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/aqcool/langextract/core"
    "github.com/aqcool/langextract/extract"
)

func main() {
    ctx := context.Background()
    
    // 输入文本
    text := "Lady Juliet gazed longingly at the stars, her heart aching for Romeo"
    
    // 提取结构化信息
    results, err := extract.Extract(
        ctx,
        text,
        extract.WithAPIKey("your-api-key"),
        extract.WithModelID("gpt-4o-mini"),
        extract.WithPromptDescription("提取人物、情感和关系。"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 打印结果
    for _, doc := range results {
        fmt.Printf("文档: %s\n", doc.DocumentID)
        for _, ext := range doc.Extractions {
            fmt.Printf("  %s: %s\n", ext.ExtractionClass, ext.ExtractionText)
        }
    }
}
```

### 使用 Few-Shot 示例提取

```go
// 定义示例来引导模型
examples := []*core.ExampleData{
    core.NewExampleData(
        "ROMEO. But soft! What light through yonder window breaks?",
        &core.Extraction{
            ExtractionClass: "character",
            ExtractionText:  "ROMEO",
            Attributes:       map[string]interface{}{"emotional_state": "wonder"},
        },
    ),
}

results, err := extract.Extract(
    ctx,
    text,
    extract.WithAPIKey("your-api-key"),
    extract.WithPromptDescription("提取人物和情感。"),
    extract.WithExamples(examples),
)
```

## 配置选项

### 模型配置

- `WithModelID(modelID string)`：设置 LLM 模型 ID（默认："gpt-4o-mini"）
- `WithAPIKey(apiKey string)`：设置 API 密钥（或使用 `OPENAI_API_KEY` 环境变量）
- `WithBaseURL(baseURL string)`：设置自定义 API 基础 URL（用于兼容 OpenAI API 的服务）
- `WithTemperature(temp float64)`：设置采样温度
- `WithMaxWorkers(workers int)`：设置并行工作线程数（默认：10）

### 提取配置

- `WithPromptDescription(desc string)`：设置提取指令（必需）
- `WithExamples(examples []*core.ExampleData)`：设置 few-shot 示例
- `WithMaxCharBuffer(chars int)`：设置最大分块大小（默认：1000）
- `WithExtractionPasses(passes int)`：设置提取轮数（默认：1）
- `WithFormatType(formatType core.FormatType)`：设置输出格式（JSON 或 YAML）

### 对齐配置

- `WithEnableFuzzyAlign(enable bool)`：启用/禁用模糊文本对齐
- `WithFuzzyThreshold(threshold float64)`：设置模糊对齐阈值（默认：0.75）

## API 参考

### 主要 API

```go
// 从文本提取结构化信息
func Extract(ctx context.Context, textOrDocuments interface{}, opts ...Option) ([]*core.AnnotatedDocument, error)

// 从文件提取
func ExtractFromFile(ctx context.Context, path string, opts ...Option) ([]*core.AnnotatedDocument, error)
```

### 核心类型

```go
// Extraction 表示单个提取实体
type Extraction struct {
    ExtractionClass  string                 // 提取类别
    ExtractionText   string                 // 提取文本
    CharInterval     *CharInterval          // 字符区间
    AlignmentStatus  AlignmentStatus        // 对齐状态
    Attributes       map[string]interface{} // 属性
}

// AnnotatedDocument 表示带提取结果的文档
type AnnotatedDocument struct {
    DocumentID   string         // 文档 ID
    Text         *string        // 文本内容
    Extractions  []*Extraction  // 提取结果列表
}

// ExampleData 表示 few-shot 示例
type ExampleData struct {
    Text        string         // 示例文本
    Extractions []*Extraction  // 示例提取结果
}
```

## 架构

```
langextract/
├── core/           # 核心数据模型和类型
│   ├── data.go     # Extraction, Document, AnnotatedDocument
│   ├── types.go    # FormatType, Constraint, ScoredOutput
│   ├── errors.go   # 自定义错误类型
│   └── tokenizer.go # 文本分词
├── providers/      # LLM 提供商接口
│   ├── base.go     # BaseLanguageModel 接口
│   └── openai/     # OpenAI 兼容提供商
├── chunking/       # 文本分块
├── prompting/      # 提示构建
├── resolver/       # 输出解析和文本对齐
├── factory/        # 模型工厂
└── extract/        # 主提取 API
```

## 环境变量

- `OPENAI_API_KEY`：OpenAI API 密钥
- `LANGEXTRACT_API_KEY`：备用 API 密钥

## 示例

查看 `examples/` 目录获取更多使用示例：

- `examples/basic/`：基础提取示例
- `examples/advanced/`：高级用法和自定义配置

## 与 Python 版本的对比

此 Go 实现保持了与 Python 版本的 API 兼容性，同时利用了 Go 的优势：

- **并发性**：原生 goroutine 支持并行处理
- **类型安全**：Go 类型系统提供强类型保证
- **性能**：编译型二进制，无运行时解释
- **部署**：单一二进制文件部署

## 许可证

Apache License 2.0

## 贡献

欢迎贡献！提交 PR 前请阅读贡献指南。

## 致谢

本项目是 Google [LangExtract](https://github.com/google/langextract) 库的 Go 语言实现。

## 免责声明

这不是 Google 官方支持的产品。使用受 Apache 2.0 许可证约束。
