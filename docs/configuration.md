# Konfiguratsiya ma’lumotnomasi

PGO MCP konfiguratsiyani faqat environment variable’lardan o‘qiydi. `.env` faylini application o‘zi avtomatik yuklamaydi.

## O‘zgaruvchilar

| O‘zgaruvchi | Majburiy | Default | Ruxsat etilgan qiymat |
| --- | --- | --- | --- |
| `DATABASE_URI` | Ha | Yo‘q | pgx qabul qiladigan PostgreSQL URI |
| `DATABASE_DISCOVERY` | Yo‘q | `false` | `true` yoki `false` |
| `DATABASE_ALLOWED` | Yo‘q | Bo‘sh | Bo‘sh, vergul bilan ajratilgan nomlar yoki aynan `*` |
| `DATABASE_MAX_POOLS` | Yo‘q | `5` | `1`–`100` |
| `MCP_AUTH_TOKEN` | Ha | Yo‘q | Kamida 32 belgi |
| `MCP_HOST` | Yo‘q | `127.0.0.1` | IP manzil; hostname qabul qilinmaydi |
| `MCP_PORT` | Yo‘q | `8001` | `1`–`65535` |
| `DB_MAX_CONNS` | Yo‘q | `5` | `1`–`100` |
| `DB_MIN_CONNS` | Yo‘q | `0` | `0`–`DB_MAX_CONNS` |
| `QUERY_TIMEOUT` | Yo‘q | `10s` | `1s`–`60s` |
| `QUERY_MAX_ROWS` | Yo‘q | `100` | `1`–`1000` |
| `QUERY_MAX_BYTES` | Yo‘q | `1048576` | `1024`–`4194304` byte |

### `DATABASE_URI`

PostgreSQL Connection Pool uchun manzil. Application startup vaqtida URI’ni parse qiladi va 5 soniyalik timeout bilan ulanishni tekshiradi.

```dotenv
DATABASE_URI=postgresql://pgo_mcp_reader:CHANGE_ME@postgres:5432/procurement?sslmode=require
```

Password ichidagi `@`, `:`, `/`, `?`, `#`, `%` kabi URI uchun maxsus belgilar percent-encoding qilinishi kerak. Masalan, `p@ss:word` qiymati `p%40ss%3Aword` ko‘rinishida yoziladi. Secret’ni repository, image layer yoki log’ga kiritmang.

URI’dagi baza default baza hisoblanadi. PostgreSQL bitta connection ichida bazani almashtirmaydi, shu sabab PGO MCP har bir tanlangan baza uchun ayni host, port, credential, TLS va runtime parametrlarini saqlagan alohida pool yaratadi. MCP input’i host, credential yoki URI’ni almashtira olmaydi.

### `DATABASE_DISCOVERY` va `DATABASE_ALLOWED`

`DATABASE_DISCOVERY=false` va bo‘sh `DATABASE_ALLOWED` mavjud single-database xatti-harakatini saqlaydi. Discovery o‘chiq bo‘lsa, nomlar ro‘yxati yoki `*` berish konfiguratsiya xatosidir.

Uchta access rejimi mavjud:

1. `DATABASE_ALLOWED=` — faqat `DATABASE_URI` dagi default baza. Bu xavfsiz default.
2. `DATABASE_ALLOWED=analytics,education,reporting` — default baza va aniq ko‘rsatilgan bazalar. Bo‘sh element, masalan `analytics,,education`, rad etiladi; takrorlar olib tashlanadi va nom registri saqlanadi.
3. `DATABASE_ALLOWED=*` — `datallowconn=true`, template bo‘lmagan va joriy rol `CONNECT` huquqiga ega barcha bazalar. `*` boshqa nom bilan aralashtirilmaydi.

Ikkinchi va uchinchi rejim uchun `DATABASE_DISCOVERY=true` bo‘lishi shart. PostgreSQL ko‘pincha `CONNECT` huquqini `PUBLIC` roliga beradi; shu sabab wildcard rejimi administrator alohida nomlamagan bazalarni ham ochishi mumkin. Production’da `PUBLIC` huquqlarini audit qiling.

### `DATABASE_MAX_POOLS`

Bir vaqtda registry saqlaydigan pool’lar soni. Default pool ham limitga kiradi. Pool birinchi muvaffaqiyatli ishlatilgandan so‘ng shutdown’gacha saqlanadi; LRU eviction yo‘q. Limit to‘lsa, yangi baza uchun aniq xato qaytariladi.

`DATABASE_MAX_POOLS=5` va `DB_MAX_CONNS=5` konfiguratsion yuqori chegara sifatida 25 ta pooled connection beradi. Bu barcha connection darhol ochilishini anglatmaydi.

### `MCP_AUTH_TOKEN`

`/mcp` endpoint’i uchun Bearer token. Majburiy va kamida 32 belgidan iborat. Server tokenlarni SHA-256 orqali solishtirishga tayyorlaydi va constant-time comparison ishlatadi.

```bash
export MCP_AUTH_TOKEN="$(openssl rand -hex 32)"
```

Tokenni URL query parametrida yubormang. Production’da secret manager yoki deployment platformasining secret injection mexanizmidan foydalaning.

### `MCP_HOST`

HTTP server bind qiladigan IP. Lokal source ishga tushirish uchun default `127.0.0.1`; container ichida odatda `0.0.0.0` kerak.

```dotenv
MCP_HOST=0.0.0.0
```

Bu qiymat hostname emas, `net.ParseIP` qabul qiladigan IP bo‘lishi shart. `0.0.0.0` serverni barcha network interface’larda tinglatadi; tashqi exposure firewall yoki port mapping bilan cheklanishi kerak.

### `MCP_PORT`

HTTP port. Default `8001`.

```dotenv
MCP_PORT=8001
```

### `DB_MAX_CONNS`

Connection Pool’dagi maksimal ulanishlar soni. Default `5`, diapazon `1`–`100`.

### `DB_MIN_CONNS`

Connection Pool saqlaydigan minimal ulanishlar soni. Default `0`; qiymat `DB_MAX_CONNS` dan katta bo‘lishi mumkin emas.

### `QUERY_TIMEOUT`

Query operatsiyalari uchun Go context timeout. Default `10s`, diapazon `1s`–`60s`. `execute_sql`, `explain_query`, Index va monitoring operatsiyalari timeout bilan cheklanadi. Ayrim service’lar PostgreSQL transaction ichida qo‘shimcha statement timeout ham o‘rnatadi.

### `QUERY_MAX_ROWS`

`execute_sql` qaytaradigan maksimal satrlar soni. Limitga yetganda `truncated` qiymati `true` bo‘ladi. Default `100`, diapazon `1`–`1000`.

### `QUERY_MAX_BYTES`

`execute_sql` satrlarining ketma-ket JSON hajmi va `explain_query` rejasi uchun qo‘llanadigan javob chegarasi. Default `1048576` byte (1 MiB), diapazon 1 KiB–4 MiB. SQL ustun nomlari va umumiy transport overhead’i `execute_sql` hisobiga kiritilmaydi.

## To‘liq namuna

```dotenv
DATABASE_URI=postgresql://pgo_mcp_reader:CHANGE_ME@postgres:5432/procurement?sslmode=require
DATABASE_DISCOVERY=true
DATABASE_ALLOWED=analytics,education,reporting
DATABASE_MAX_POOLS=5
DB_MAX_CONNS=5
DB_MIN_CONNS=0

MCP_HOST=0.0.0.0
MCP_PORT=8001
MCP_AUTH_TOKEN=REPLACE_WITH_A_RANDOM_64_CHARACTER_HEX_TOKEN

QUERY_TIMEOUT=10s
QUERY_MAX_ROWS=100
QUERY_MAX_BYTES=1048576
```

TLS database konfiguratsiyasini deployment muhitingizga moslang. Ishonchsiz network’da `sslmode=disable` ishlatmang.
