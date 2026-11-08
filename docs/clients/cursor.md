# Cursor bilan ulanish

Cursor rasmiy hujjatlariga ko‘ra remote Streamable HTTP serverlar, `url` va custom `headers` konfiguratsiyasini qo‘llaydi. Quyidagi misol shu rasmiy formatga asoslangan; ushbu repository doirasida Cursor bilan end-to-end sinovdan o‘tkazilmagan.

Rasmiy manba: [Cursor MCP hujjatlari](https://cursor.com/docs/mcp).

## Konfiguratsiya joyi

- Faqat bitta project uchun: `.cursor/mcp.json`.
- Barcha projectlar uchun shaxsiy konfiguratsiya: `~/.cursor/mcp.json`.

Tokenni project repository’siga commit qilmaslik uchun global shaxsiy fayldan foydalanish xavfsizroq.

```json
{
  "mcpServers": {
    "pgo-mcp": {
      "url": "http://127.0.0.1:8001/mcp",
      "headers": {
        "Authorization": "Bearer REPLACE_WITH_YOUR_TOKEN"
      }
    }
  }
}
```

`REPLACE_WITH_YOUR_TOKEN` ni serverdagi `MCP_AUTH_TOKEN` bilan almashtiring. Tokenli faylni source control’ga qo‘shmang va filesystem permission’larini cheklang. Cursor versiyangiz environment substitution’ni qo‘llashini alohida tekshirmasdan `${...}` sintaksisiga tayanmang.

Fayl saqlangach Cursor’ni qayta ishga tushiring yoki MCP serverni UI orqali restart qiling.

## Tekshirish

Cursor’da MCP sozlamalarini oching va `pgo-mcp` serveri ulanganini tekshiring. Agent’dan `list_schemas` ni ishlatishni so‘rang. Tool ro‘yxatda ko‘ringani bilan cheklanmay, javobda real schema ro‘yxati kelganini tekshiring.

## Muammolarni aniqlash

- Cursor Output panelida `MCP Logs` kanalini oching.
- `401 Unauthorized` bo‘lsa header nomi aynan `Authorization`, qiymati `Bearer <token>` ekanini tekshiring.
- Duplicate server yozuvlarini `.cursor/mcp.json` va `~/.cursor/mcp.json` ichidan qidiring.
- URL `http://127.0.0.1:8001/mcp` bo‘lishi kerak; `/health` MCP endpoint emas.
- Cursor remote workspace yoki container ichida ishlasa, uning `127.0.0.1` manzili macOS host’ini anglatmasligi mumkin. Reachable URL ishlating.

PGO MCP OAuth authorization server emas. Cursor OAuth dialogini ko‘rsatsa, custom static `Authorization` header konfiguratsiyasi yuklanganini va eski server yozuvi qolmaganini tekshiring.
