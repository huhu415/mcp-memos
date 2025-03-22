# MCP-Memos

MCP-Memos is an memo tool based on [MCP](https://modelcontextprotocol.io/introduction) that allows users to record and retrieve text information.

It's perfect for developers to quickly save and find information in their workflow without switching to other applications.

## Tool Documentation

| Tool Name | Description | Input Parameters |
|-----------|-------------|------------------|
|store_memo| Save important text information and add tags for easy retrieval  later| `tag`:keyword or tag or description <br> `content`: that you want to save    |
|retrieve_memo| Retrieve previously saved text content based on keywords | `text` keyword or tag or description |

## Usage
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
          "LLM_BASE_URL": "xxxxx" # optional, default is https://api.anthropic.com
          "ANTHROPIC_MODEL": "xxxxx" # optional, default is claude-3-7-sonnet-20250219
        }
      }
    }
}
```

### Record information
Sometimes when we're developing, we need to record some information.
1. Don't want to open a note-taking app to record, it's too cumbersome
2. If you've opened a note-taking app to record, it's difficult to find the information later


### Record information
You can use MCP-Memos to record information by saying:
```
Please record this for me. This is {description}

{content}
```

### Retrieve information
When you need it later, just say:
```
Please help me find the records about {description}
```

> [!NOTE]
> The description can be different each time, as long as the description is roughly the same thing
