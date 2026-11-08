# Claude Code bilan ulanish

PGO MCP Streamable HTTP va statik Bearer token ishlatadi. Claude Code rasmiy hujjatlariga ko‘ra remote HTTP server va custom `Authorization` header’ni native qo‘llaydi. Quyidagi konfiguratsiya rasmiy imkoniyatga asoslangan; ushbu repository doirasida Claude Code bilan end-to-end sinovdan o‘tkazilmagan.

Rasmiy manba: [Claude Code MCP qo‘llanmasi](https://code.claude.com/docs/en/mcp).

## Talablar

- PGO MCP `http://127.0.0.1:8001/mcp` da ishlayotgan bo‘lishi;
- Claude Code o‘rnatilgan bo‘lishi;
- `MCP_AUTH_TOKEN` server tokeni bilan bir xil bo‘lishi.

## `.mcp.json` orqali

Project ildizida `.mcp.json` yarating:

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

Tokenni Claude Code ishga tushiriladigan muhitga bering:

```bash
export MCP_AUTH_TOKEN=REPLACE_WITH_YOUR_TOKEN
claude
```

`.mcp.json` ni team bilan ulashish mumkin, chunki unda secret qiymat emas, environment reference bor. Real tokenni commit qilmang.

## CLI orqali

Claude Code remote HTTP serverni CLI orqali ham qo‘sha oladi:

```bash
claude mcp add --transport http pgo-mcp http://127.0.0.1:8001/mcp \
  --header "Authorization: Bearer ${MCP_AUTH_TOKEN}"
```

Bu usul tokenni Claude konfiguratsiyasida saqlashi mumkin. Shared kompyuterda konfiguratsiya va shell history siyosatini tekshiring; environment reference’li `.mcp.json` afzal.

## Tekshirish

Claude Code ichida `/mcp` orqali server holatini ko‘ring. `list_schemas` vositasini ishlatishni so‘rang va natija haqiqiy PostgreSQL schemalarini qaytarganini tekshiring. Faqat vosita nomi paydo bo‘lishi yetarli emas; real tool call natijasini ham ko‘ring.

## Muammolarni aniqlash

- `401 Unauthorized`: token serverdagi `MCP_AUTH_TOKEN` bilan bir xil ekanini va environment Claude process’iga meros bo‘lganini tekshiring.
- Server konfiguratsiyada yo‘q: `type` qiymati `http`, URL oxiri `/mcp` ekanini tekshiring.
- Connection rad etildi: PGO MCP va `/health` endpoint holatini tekshiring.
- Remote HTTPS server: sertifikat chain’i Claude Code ishlayotgan host tomonidan ishonchli ekanini tekshiring.

Claude Desktop yoki claude.ai konfiguratsiyasi Claude Code’dan farq qilishi mumkin. Ushbu sahifadagi misol aynan Claude Code uchun.
