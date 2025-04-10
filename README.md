# MCP-Memos

MCP-Memos is an memo tool based on [MCP](https://modelcontextprotocol.io/introduction) that allows users to record and retrieve text information.

It's perfect for developers to quickly save and find information in their workflow(may be cursor, etc) without switching to other applications.

## 🔍 Advanced LLM-Powered Search

MCP-Memos uses large language models for retrieval, providing the **most powerful fuzzy search capability available**:

- **Semantic understanding**: Find content based on meaning, not just keywords
- **Context-aware**: Understands what you're looking for even with incomplete descriptions
- **Natural language queries**: Search as you would ask a human, no special syntax needed
- **Conceptual matching**: Retrieves information by understanding concepts, not just text matching

Unlike traditional vector or text-based fuzzy search, MCP-Memos leverages the full power of LLMs to truly understand your retrieval intent, making it the most effective information retrieval approach available today.

## Tool Documentation

| Tool Name | Description | Input Parameters |
|-----------|-------------|------------------|
|store_memo| Save important text information and add tags for easy retrieval  later| `tag`:keyword or tag or description <br> `content`: that you want to save    |
|retrieve_memo| Retrieve previously saved text content based on keywords | `text` keyword or tag or description |

## How to Use
### Installation
Download the mcp-memos binary file according to your computer's architecture from [releases](https://github.com/huhu415/mcp-memos/releases)

### Configuration
Add MCP-Memos to the macPilotCli configuration file
```
{
    "mcpServers": {
      "MCP-Memos":{
        "command": "path/to/mcp-memos",
        "env": {
          "LLM_TOKEN": "xxxxxx",
          "LLM_BASE_URL": "xxxxxxxx/v1",
          "ANTHROPIC_MODEL": "xxxxx"
        }
      }
    }
}
```

### Basic Usage

### Record information
When you need to save information during development without switching to another app:

```
record this memo. its {description}

{content}
```

### Retrieve information
When you need to find previously saved information:

```
Find this memo about {description}
```

> [!NOTE]
> Descriptions can vary as MCP-Memo's LLM understands concepts, not just keywords, matching similar content regardless of wording.
>
> LLM_BASE_URL optional, default is https://api.anthropic.com
>
> ANTHROPIC_MODEL optional, default is claude-3-7-sonnet-20250219
>
> Will accelerate the implementation of the [sampling](https://spec.modelcontextprotocol.io/specification/2025-03-26/client/sampling/) feature, which will eliminate the need to configure all environment parameters
