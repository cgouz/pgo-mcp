# O‘zgarishlar tarixi

Ushbu fayl PGO MCP release’laridagi foydalanuvchiga ta’sir qiladigan o‘zgarishlarni qayd etadi.

## [0.2.0]

### Qo‘shildi

- Bitta PostgreSQL server va `DATABASE_URI` orqali bir nechta bazada ishlash imkoniyati.
- PostgreSQL katalogi va `CONNECT` huquqi asosidagi database discovery.
- Ruxsat etilgan va default bazalarni qaytaradigan `list_databases` MCP vositasi.
- Avvalgi to‘qqizta MCP vositasi uchun ixtiyoriy `database` parametri.
- Har bir baza uchun mustaqil pool yaratadigan concurrency-safe Database Registry.
- `DATABASE_MAX_POOLS` orqali boshqariladigan pool chegarasi.
- Bo‘sh qiymat, aniq nomlar ro‘yxati va `*` wildcard rejimlariga ega administrator allowlist’i.

### Yaxshilandi

- Parallel MCP chaqiruvlarida database tanlovining o‘zaro izolyatsiyasi.
- Bir database uchun concurrent so‘rovlarda pool’ni takror yaratmaslik va mavjud pool’ni qayta ishlatish.
- Ikki ruxsatli test bazasi hamda `CONNECT` huquqi yo‘q baza bilan PostgreSQL integration testlari.
- MCP tool xatolarini test loglarida o‘qiladigan matn yoki JSON ko‘rinishida chiqarish.
- SQL, HTTP request, MCP session va server resource limitlari bo‘yicha xavfsizlik tekshiruvlari.

### Moslik

- Yangi database variable’lari berilmasa, mavjud single-database konfiguratsiyasi avvalgidek ishlaydi.
- Avvalgi to‘qqizta MCP vositasi, ularning nomlari va output tuzilmalari saqlangan.
- `database` parametri berilmasa yoki bo‘sh bo‘lsa, `DATABASE_URI` dagi default baza ishlatiladi.

## [0.1.0]

- PGO MCP’ning dastlabki release’i: to‘qqizta MCP vositasi, Streamable HTTP, Bearer authentication, PostgreSQL connection pool va read-only query tahlili.
