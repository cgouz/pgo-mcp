# Visual Studio Code bilan ulanish

Visual Studio Code rasmiy MCP konfiguratsiyasi Streamable HTTP va custom HTTP header’larni qo‘llaydi. Quyidagi Local harness misoli `${input:...}` orqali tokenni interaktiv so‘raydi. Ushbu repository doirasida VS Code bilan end-to-end sinovdan o‘tkazilmagan.

Rasmiy manbalar:

- [MCP server qo‘shish](https://code.visualstudio.com/docs/agent-customization/mcp-servers)
- [MCP konfiguratsiya ma’lumotnomasi](https://code.visualstudio.com/docs/agents/reference/mcp-configuration)

## Workspace konfiguratsiyasi

`.vscode/mcp.json` faylini yarating:

```json
{
  "servers": {
    "pgo-mcp": {
      "type": "http",
      "url": "http://127.0.0.1:8001/mcp",
      "headers": {
        "Authorization": "Bearer ${input:pgo-mcp-token}"
      }
    }
  },
  "inputs": [
    {
      "type": "promptString",
      "id": "pgo-mcp-token",
      "description": "PGO MCP Bearer token",
      "password": true
    }
  ]
}
```

VS Code serverni boshlaganda tokenni kiriting. Real tokenni JSON faylga yozmang yoki commit qilmang.

## Portable Copilot konfiguratsiyasi

Workspace ildizidagi `.mcp.json` portable formatda `inputs` ishlamaydi. Environment reference’dan foydalaning:

```json
{
  "mcpServers": {
    "pgo-mcp": {
      "type": "http",
      "url": "http://127.0.0.1:8001/mcp",
      "headers": {
        "Authorization": "Bearer ${MCP_AUTH_TOKEN}"
      }
    }
  }
}
```

`MCP_AUTH_TOKEN` agent host process’i ishga tushadigan muhitda mavjud bo‘lishi kerak. Oddiy terminalda keyinroq export qilish allaqachon ishlayotgan VS Code process’ini yangilamaydi; VS Code’ni tegishli environment bilan qayta ishga tushiring.

Local harness va Copilot harness konfiguratsiya formatlari hamda secret input qo‘llovi farq qiladi. Konfiguratsiyani bir formatdan boshqasiga ko‘r-ko‘rona ko‘chirmang.

## Tekshirish

Command Palette’dan `MCP: List Servers` yoki `Chat: Open Customizations` → `MCP Servers` ni oching. `pgo-mcp` ni boshlang, tool ro‘yxatini ko‘ring va `list_schemas` ni real chaqirib tekshiring.

## Muammolarni aniqlash

- `401 Unauthorized`: token va `Authorization` header’ni tekshiring.
- Server ishlamaydi: `MCP: List Servers` orqali `Show Output` ni oching.
- Copilot harness serverni ko‘rmaydi: portable `.mcp.json` formatidan va environment reference’dan foydalaning; interactive `inputs` forwarding qilinmaydi.
- Remote workspace yoki Dev Container’da URL o‘sha environment’dan reachable bo‘lishi kerak.
- Log yoki issue yuborishdan oldin token va authorization header’ni redaction qiling.
