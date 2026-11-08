# MCP vositalari ma’lumotnomasi

Barcha vositalar `http://127.0.0.1:8001/mcp` Streamable HTTP endpoint’i orqali chaqiriladi va to‘g‘ri Bearer token talab qiladi. Misollardagi `arguments` MCP `tools/call` argument obyektini bildiradi.

`list_databases` dan tashqari barcha vositalar ixtiyoriy `database` string parametrini qabul qiladi. Parametr berilmasa yoki bo‘sh bo‘lsa, `DATABASE_URI` dagi default baza ishlatiladi. Qiymat faqat joriy tool chaqiruviga tegishli; u global holatni o‘zgartirmaydi. Nom `DATABASE_ALLOWED` siyosati va PostgreSQL `CONNECT` huquqidan o‘tishi shart.

## `list_databases`

Faol konfiguratsiya ruxsat bergan, template bo‘lmagan va PostgreSQL roli `CONNECT` qila oladigan bazalarni barqaror tartibda qaytaradi. Bu vosita yangi pool’larni avtomatik ochmaydi va connection URI yoki credential’larni ko‘rsatmaydi.

Input parametrlari yo‘q.

```json
{
  "name": "list_databases",
  "arguments": {}
}
```

```json
{
  "default": "procurement",
  "databases": [
    {"name": "analytics", "default": false},
    {"name": "procurement", "default": true}
  ]
}
```

Discovery o‘chiq yoki allowlist bo‘sh bo‘lsa, faqat default baza qaytariladi.

## `list_schemas`

Ulangan PostgreSQL bazasidagi schemalarni alifbo tartibida qaytaradi.

`database` dan boshqa input parametrlari yo‘q.

```json
{
  "name": "list_schemas",
  "arguments": {}
}
```

Output:

```json
{
  "schemas": ["information_schema", "public"]
}
```

`schemas` — `string[]`. Natija ulangan rolning PostgreSQL kataloglari orqali ko‘ra oladigan ma’lumotlariga bog‘liq.

## `list_objects`

Berilgan schema ichidagi ruxsat etilgan obyektlarni qaytaradi.

| Parametr | Tur | Majburiy | Tavsif |
| --- | --- | --- | --- |
| `schema_name` | `string` | Ha | PostgreSQL schema nomi. |
| `object_type` | `string` | Yo‘q | `table`, `view`, `sequence` yoki `extension`; default `table`. |

```json
{
  "name": "list_objects",
  "arguments": {
    "schema_name": "public",
    "object_type": "table"
  }
}
```

```json
{
  "schema_name": "public",
  "object_type": "table",
  "objects": ["procurements"]
}
```

Output’dagi `objects` — `string[]`. Table va view uchun `USAGE` hamda `SELECT`, sequence uchun `USAGE` va `SELECT` privilege tekshiriladi. Noma’lum `object_type` xato qaytaradi.

## `get_object_details`

Table, view, sequence yoki extension tafsilotlarini beradi.

| Parametr | Tur | Majburiy | Tavsif |
| --- | --- | --- | --- |
| `schema_name` | `string` | Ha | Schema nomi. |
| `object_name` | `string` | Ha | Obyekt nomi. |
| `object_type` | `string` | Yo‘q | `table`, `view`, `sequence` yoki `extension`; default `table`. |

```json
{
  "name": "get_object_details",
  "arguments": {
    "schema_name": "public",
    "object_name": "procurements",
    "object_type": "table"
  }
}
```

Table uchun namuna:

```json
{
  "schema_name": "public",
  "object_name": "procurements",
  "object_type": "table",
  "columns": [
    {
      "name": "id",
      "data_type": "bigint",
      "nullable": false,
      "default": "nextval('procurements_id_seq'::regclass)",
      "primary_key": true
    }
  ],
  "constraints": [
    {
      "name": "procurements_pkey",
      "type": "PRIMARY KEY",
      "definition": "PRIMARY KEY (id)"
    }
  ],
  "indexes": [
    {
      "name": "procurements_pkey",
      "definition": "CREATE UNIQUE INDEX procurements_pkey ON public.procurements USING btree (id)"
    }
  ]
}
```

Asosiy output maydonlari:

- `schema_name`, `object_name`, `object_type` — `string`;
- `columns[]` — `name`, `data_type`, `nullable`, ixtiyoriy `default`, `primary_key`;
- `constraints[]` — `name`, `type`, `definition`;
- `indexes[]` — `name`, `definition`;
- view uchun `definition`;
- sequence uchun `metadata`: `data_type`, `start`, `increment`, `min`, `max`, `cache`, `cycle`;
- extension uchun `metadata`: `version`, `description`.

Faqat ulangan rol tegishli privilege’ga ega obyektlar ochiladi.

## `execute_sql`

SQL so‘rovini PostgreSQL read-only transaction ichida bajaradi. Ustun tartibi va `NULL` qiymatlar saqlanadi; `NULL` JSON’da `null` bo‘ladi, boshqa SQL qiymatlar serverdagi joriy string ko‘rinishiga o‘tkaziladi. `bytea` qiymati `\x` prefiksli hex, vaqt esa RFC3339Nano formatida qaytariladi.

| Parametr | Tur | Majburiy | Tavsif |
| --- | --- | --- | --- |
| `sql` | `string` | Ha | Bajariladigan SQL. Bo‘sh qiymat qabul qilinmaydi. |

```json
{
  "name": "execute_sql",
  "arguments": {
    "database": "analytics",
    "sql": "SELECT id, status FROM public.procurements ORDER BY id LIMIT 2"
  }
}
```

```json
{
  "columns": ["id", "status"],
  "rows": [
    ["1", "active"],
    ["2", null]
  ],
  "row_count": 2,
  "truncated": false
}
```

`rows` — nullable string kataklardan iborat ikki o‘lchamli array. `QUERY_MAX_ROWS` yoki `QUERY_MAX_BYTES` chegarasi kesilsa, `truncated=true` qaytadi. Transaction read-only va PostgreSQL statement timeout bilan cheklangan. Shunga qaramay, ruxsat berilgan function’lar tashqi side effect yaratishi mumkin; [xavfsizlik tavsiyalariga](security.md) amal qiling.

## `explain_query`

PostgreSQL Query bajarilish rejasini `FORMAT JSON, VERBOSE TRUE` bilan qaytaradi.

| Parametr | Tur | Majburiy | Tavsif |
| --- | --- | --- | --- |
| `sql` | `string` | Ha | Tahlil qilinadigan SQL; maksimal 64 KiB. |
| `analyze` | `boolean` | Yo‘q | Default `false`; `true` bo‘lsa Query amalda bajariladi va `BUFFERS` statistikasi yig‘iladi. |

```json
{
  "name": "explain_query",
  "arguments": {
    "sql": "SELECT * FROM public.procurements WHERE status = 'active'",
    "analyze": false
  }
}
```

```json
{
  "analyze": false,
  "plan": [
    {
      "Plan": {
        "Node Type": "Seq Scan",
        "Relation Name": "procurements",
        "Schema": "public",
        "Plan Rows": 1,
        "Total Cost": 20.62
      }
    }
  ]
}
```

`plan` PostgreSQL yaratgan JSON obyektlar array’idir va reja maydonlari PostgreSQL versiyasi hamda Query’ga qarab o‘zgaradi. `analyze=true` Query’ni real bajaradi; faqat xavfsiz, kutilgan so‘rovlar bilan ishlating.

## `analyze_query_indexes`

Bitta SQL Query uchun `EXPLAIN` rejasidagi scan node’larini, aloqador jadvallar va mavjud indekslarni tahlil qiladi. Query’ni `ANALYZE` bilan bajarmaydi va indeks yaratmaydi yoki o‘chirmaydi.

| Parametr | Tur | Majburiy | Tavsif |
| --- | --- | --- | --- |
| `sql` | `string` | Ha | Tahlil qilinadigan SQL. |

```json
{
  "name": "analyze_query_indexes",
  "arguments": {
    "sql": "SELECT * FROM public.procurements WHERE status = 'active'"
  }
}
```

```json
{
  "estimated_cost": 20.62,
  "scans": [
    {
      "node_type": "Seq Scan",
      "schema": "public",
      "relation": "procurements",
      "filter": "(status = 'active'::text)",
      "estimated_rows": 1,
      "total_cost": 20.62
    }
  ],
  "tables": [
    {
      "schema": "public",
      "table": "procurements",
      "size_bytes": 8192,
      "estimated_rows": 2,
      "indexes": []
    }
  ],
  "findings": [
    {
      "severity": "info",
      "message": "No obvious index-related warning detected"
    }
  ]
}
```

`scans[]` ichida `node_type`, ixtiyoriy `schema`, `relation`, `index_name`, `condition`, `filter`, shuningdek `estimated_rows` va `total_cost` bo‘ladi. `tables[].indexes[]` har bir Index uchun `name`, `definition`, `unique`, `valid`, `ready` qiymatlarini beradi. Natija tavsiya xarakterida; DDL avtomatik bajarilmaydi.

## `analyze_workload_indexes`

Butun bazadagi workload Query va Index usage statistikasini tahlil qiladi. `pg_stat_statements` extension’i talab qilinadi.

| Parametr | Tur | Majburiy | Tavsif |
| --- | --- | --- | --- |
| `limit` | `integer` | Yo‘q | Default `20`, diapazon `1`–`50`. |

```json
{
  "name": "analyze_workload_indexes",
  "arguments": {
    "limit": 10
  }
}
```

```json
{
  "database": "procurement",
  "queries": [
    {
      "query_id": "12345",
      "query": "SELECT status FROM public.procurements",
      "calls": 12,
      "total_exec_ms": 8.4,
      "mean_exec_ms": 0.7,
      "rows": 24,
      "shared_hits": 20,
      "shared_reads": 1,
      "temp_written": 0
    }
  ],
  "indexes": [],
  "findings": [
    {
      "severity": "info",
      "message": "No obvious workload or index warning detected"
    }
  ]
}
```

`queries[]` faqat joriy PostgreSQL user’iga tegishli `pg_stat_statements` yozuvlarini oladi. `indexes[]` maydonlari: `schema`, `table`, `index`, `scans`, `size_bytes`, `primary`, `unique`. Vosita faqat tahlil qiladi, Index DDL bajarmaydi.

`analyze_query_indexes` bitta berilgan Query’ning planner rejasini tekshiradi; `analyze_workload_indexes` esa vaqt davomida yig‘ilgan database workload va Index statistikalarini ko‘radi.

## `analyze_db_health`

Database hajmi, connection, transaction, cache, deadlock, temporary file va eng ko‘p dead tuple’ga ega jadvallarni tahlil qiladi.

Input parametrlari yo‘q.

```json
{
  "name": "analyze_db_health",
  "arguments": {}
}
```

```json
{
  "database": "procurement",
  "size_bytes": 10485760,
  "collected_at": "2026-11-08T10:00:00Z",
  "connections": {
    "database_active": 1,
    "database_idle": 2,
    "database_other": 0,
    "server_total": 3,
    "max_connections": 100,
    "usage_percent": 3
  },
  "transactions": {
    "commits": 100,
    "rollbacks": 2,
    "rollback_ratio": 1.9607843137254901
  },
  "cache": {
    "blocks_hit": 5000,
    "blocks_read": 100,
    "hit_ratio": 98.0392156862745
  },
  "deadlocks": 0,
  "temp_files": 0,
  "temp_bytes": 0,
  "tables": [],
  "findings": [
    {
      "severity": "info",
      "category": "general",
      "message": "No immediate issue detected by the configured health checks"
    }
  ]
}
```

`stats_reset`, ratio qiymatlari va table’dagi `last_autovacuum`/`last_autoanalyze` ma’lumot bo‘lmasa tushirib qoldirilishi mumkin. `tables` eng ko‘p dead tuple bo‘yicha 10 tagacha table beradi. To‘liq server statistikasi uchun odatda `pg_read_all_stats` membership kerak.

## `get_top_queries`

`pg_stat_statements` yozuvlarini tanlangan mezon bo‘yicha saralaydi.

| Parametr | Tur | Majburiy | Tavsif |
| --- | --- | --- | --- |
| `sort_by` | `string` | Yo‘q | `total_time`, `mean_time`, `calls` yoki `rows`; default `total_time`. |
| `limit` | `integer` | Yo‘q | Default `10`, diapazon `1`–`50`. |

```json
{
  "name": "get_top_queries",
  "arguments": {
    "sort_by": "calls",
    "limit": 10
  }
}
```

```json
{
  "database": "procurement",
  "sort_by": "calls",
  "count": 1,
  "queries": [
    {
      "query_id": "12345",
      "query": "SELECT status FROM public.procurements",
      "calls": 12,
      "total_exec_ms": 8.4,
      "mean_exec_ms": 0.7,
      "min_exec_ms": 0.4,
      "max_exec_ms": 1.1,
      "rows": 24,
      "shared_hits": 20,
      "shared_reads": 1,
      "stats_since": "2026-11-08T09:00:00Z"
    }
  ]
}
```

`pg_stat_statements` extension’i majburiy. Joriy user o‘z Query’larini ko‘radi; `pg_read_all_stats` a’zoligi bo‘lsa boshqa user statistikasi ham ko‘rinishi mumkin. Query matni 2048 belgigacha qisqartiriladi.

## Permission va xavfsizlik xulosasi

- Barcha vositalar Bearer token talab qiladi.
- Catalog vositalari PostgreSQL privilege tekshiruvlaridan foydalanadi.
- Query vositalari read-only transaction ishlatadi, ammo database function huquqlari alohida nazorat qilinishi kerak.
- Monitoring vositalari uchun `pg_stat_statements` va ayrim holatlarda `pg_read_all_stats` talab qilinadi.
- Hech bir Index tahlili vositasi schema yoki indekslarni avtomatik o‘zgartirmaydi.
