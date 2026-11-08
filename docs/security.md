# Xavfsizlik

PGO MCP database metadata va Query natijalarini modelga uzatadi. Uni oddiy monitoring endpoint’i emas, himoyalangan database client sifatida ko‘ring.

## Himoya qatlamlari

Application quyidagi cheklovlarni qo‘llaydi:

- `/mcp` uchun majburiy Bearer authentication;
- PostgreSQL Connection Pool orqali alohida login roli;
- administrator allowlist’i va har bir baza uchun PostgreSQL `CONNECT` tekshiruvi;
- Query operatsiyalari uchun read-only transaction;
- Go context timeout va PostgreSQL statement timeout;
- `execute_sql` uchun satr va JSON hajmi limitlari;
- `explain_query` rejasi uchun hajm limiti;
- catalog so‘rovlarida PostgreSQL privilege tekshiruvi;
- Docker image’da non-root UID, minimal runtime va static binary.

Bu qatlamlar bir-birini to‘ldiradi. Ulardan biri qolganlarini almashtirmaydi.

## Read-only PostgreSQL roli

Quyidagi namuna bitta bazaning `public` schemasini o‘qish uchun rol yaratadi. Password’ni secret manager orqali kuchli qiymatga almashtiring.

```sql
CREATE ROLE pgo_mcp_reader
    LOGIN
    PASSWORD 'REPLACE_WITH_A_STRONG_PASSWORD'
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOREPLICATION;

GRANT CONNECT ON DATABASE procurement TO pgo_mcp_reader;
GRANT CONNECT ON DATABASE analytics TO pgo_mcp_reader;
GRANT USAGE ON SCHEMA public TO pgo_mcp_reader;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO pgo_mcp_reader;
GRANT SELECT ON ALL SEQUENCES IN SCHEMA public TO pgo_mcp_reader;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT ON TABLES TO pgo_mcp_reader;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT ON SEQUENCES TO pgo_mcp_reader;
```

`ALTER DEFAULT PRIVILEGES` faqat uni bajargan owner kelajakda yaratadigan obyektlarga ta’sir qiladi. Bir nechta owner bo‘lsa, har biri uchun alohida sozlang.

Monitoring vositalariga to‘liq statistika kerak bo‘lsa:

```sql
GRANT pg_read_all_stats TO pgo_mcp_reader;
```

Bu membership server bo‘ylab kengroq statistik ko‘rinish beradi. Faqat zarur bo‘lsa bering va tashkilot siyosatiga muvofiq audit qiling.

## Read-only cheklovining chegarasi

Read-only transaction odatiy `INSERT`, `UPDATE`, `DELETE`, DDL va shunga o‘xshash database yozuvlarini to‘sadi. Lekin `SELECT` orqali chaqiriladigan ayrim function’lar:

- tashqi HTTP yoki filesystem operatsiyasini bajarishi;
- extension orqali tashqi tizimga ta’sir qilishi;
- advisory lock yoki boshqa session holatini o‘zgartirishi;
- noto‘g‘ri yozilgan security-definer function orqali ortiqcha huquq ishlatishi mumkin.

Shuning uchun `SELECT` privilege yakka o‘zi to‘liq sandbox emas. `PUBLIC` va reader rolidan xavfli function’lar uchun `EXECUTE` huquqini olib tashlang, kerakli function’larni allowlist tamoyili bilan bering, extension va security-definer function’larni audit qiling.

`explain_query` uchun `analyze=true` real Query’ni bajaradi. Read-only transaction ichida bo‘lsa ham, uni faqat ishonchli Query bilan ishlating.

## Bearer authentication

`MCP_AUTH_TOKEN` kamida 32 belgili bo‘lishi shart. 32 random byte’dan hex token yaratish:

```bash
openssl rand -hex 32
```

Tokenni:

- repository yoki Docker image’ga kiritmang;
- URL query parametrida yubormang;
- log, screenshot yoki issue’da ko‘rsatmang;
- har bir muhit uchun alohida yarating;
- sizib chiqsa darhol almashtiring.

Middleware aynan bitta `Authorization: Bearer <token>` header qabul qiladi va qiymatni constant-time usulda taqqoslaydi.

## Network va TLS

Bearer authentication shifrlash emas. Oddiy HTTP orqali yuborilgan token network’da ko‘rinishi mumkin. Lokal deployment’da portni `127.0.0.1` ga bind qiling. Tashqi client uchun:

- TLS termination qiladigan ishonchli reverse proxy ishlating;
- faqat HTTPS endpoint e’lon qiling;
- firewall yoki private network bilan source’larni cheklang;
- proxy’da request size, timeout va rate limit qo‘llang;
- `/mcp` header’ini saqlab uzating, lekin log’larda redaction qiling.

`/health` application darajasida ochiq. Unda secret ma’lumot yo‘q, ammo tashqi exposure kerak bo‘lmasa uni proxy yoki firewall bilan cheklang.

## Query limitlari

- `QUERY_TIMEOUT`: 1–60 soniya, default `10s`.
- `QUERY_MAX_ROWS`: 1–1000, default `100`.
- `QUERY_MAX_BYTES`: 1 KiB–4 MiB, default 1 MiB.

Limitlar resurs sarfini kamaytiradi, lekin qimmat Query’ni butunlay zararsiz qilmaydi. PostgreSQL rol darajasida `statement_timeout`, `lock_timeout`, `idle_in_transaction_session_timeout` va zarur resource governance sozlamalarini ham ko‘rib chiqing.

## PostgreSQL transport xavfsizligi

Production URI’da database TLS’dan foydalaning va sertifikat tekshiruvini deployment siyosatiga mos sozlang. Ishonchsiz network’da `sslmode=disable` ishlatmang. Database password’ini URI bilan environment’da saqlash process inspection xavfini keltirishi mumkin; platformaning secret injection yoki PostgreSQL service file imkoniyatlarini baholang.

## Docker xavfsizligi

Rasmiy image:

- UID/GID `65532` bilan non-root ishlaydi;
- Go toolchain yoki source kodni runtime stage’ga kiritmaydi;
- CA sertifikatlarini saqlaydi;
- Compose’da `read_only`, `cap_drop: ALL` va `no-new-privileges` bilan ishlatilishi mumkin.

Image tag’ni production’da aniq versiya yoki digest bilan pin qiling:

```bash
docker pull ghcr.io/cgouz/pgo-mcp:0.2.0
```

## MCP access control

Hozirgi server bitta global Bearer token ishlatadi; per-user identity, per-tool yoki per-database authorization yoxud OAuth scope modeli yo‘q. Token egasi PGO MCP konfiguratsiyasi ruxsat bergan barcha bazalarda o‘nta vositadan foydalanishi mumkin. Ko‘p foydalanuvchili muhitda identity-aware reverse proxy, alohida instance/token va audit log strategiyasini qo‘llang.

`DATABASE_ALLOWED=*` faqat ishonchli muhit uchun explicit opt-in hisoblanadi. PostgreSQL `PUBLIC` roliga default `CONNECT` bergan bo‘lishi mumkin, shuning uchun wildcard kutilganidan ko‘proq bazani ko‘rsatishi ehtimoli bor. `pg_database`, `PUBLIC` privilege’lari va reader rol grantlarini muntazam audit qiling.
